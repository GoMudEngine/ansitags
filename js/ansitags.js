/**
 * ansitags.js
 *
 * Parses strings containing <ansi> tags and converts them to HTML <span> tags
 * with inline color styles. This is a JavaScript port of the ansitags Go library,
 * supporting the same tag format and producing identical HTML output.
 *
 * Tag format:
 *   <ansi fg="{color}" bg="{color}">text</ansi>
 *
 *   Colors may be specified as a 256-color ANSI code (0-255) or a named alias.
 *   Built-in named colors: black, red, green, yellow, blue, magenta, cyan, white,
 *   and their -bold variants (e.g. red-bold, blue-bold).
 *
 * Browser usage (include the script directly in a web page):
 *
 *   <script src="ansitags.js"></script>
 *   <script>
 *     document.getElementById('output').innerHTML =
 *       ansitags.parse('<ansi fg="red">hello</ansi> world');
 *   </script>
 *
 * Node.js usage:
 *
 *   const { parse, loadAliases, setAlias, setAliases, rgb } = require('./ansitags');
 *
 *   // Basic parsing - outputs HTML spans
 *   parse('<ansi fg="red">hello</ansi>');
 *   // => '<span style="color:#800000;">hello</span>'
 *
 *   // Foreground and background
 *   parse('<ansi fg="blue" bg="yellow">hello</ansi>');
 *   // => '<span style="color:#000080;background-color:#808000;">hello</span>'
 *
 *   // Numeric 256-color codes
 *   parse('<ansi fg="196">hello</ansi>');
 *   // => '<span style="color:#ff0000;">hello</span>'
 *
 *   // Nested tags - inner tag inherits unspecified properties from parent
 *   parse('<ansi fg="green">outer <ansi fg="red">inner</ansi> outer</ansi>');
 *
 *   // Strip all tags, leaving only the text content
 *   parse('<ansi fg="red">hello</ansi>', { stripTags: true });
 *   // => 'hello'
 *
 *   // Monochrome - strip color, keep tag structure
 *   parse('<ansi fg="red">hello</ansi>', { monochrome: true });
 *   // => '<span>hello</span>'
 *
 *   // Load custom color aliases (flat key:value object)
 *   loadAliases({ date: 207, username: 195, highlight: 'yellow' });
 *   parse('<ansi fg="date">2026-04-22</ansi>');
 *
 *   // Look up RGB values for any ANSI 256 color code
 *   rgb(196); // => { r: 255, g: 0, b: 0, hex: 'ff0000' }
 */

'use strict';

// ANSI 256-color palette: RGB values for codes 0-255
const ansi256 = (() => {
  const palette = new Array(256);

  const base16 = [
    [0, 0, 0], [128, 0, 0], [0, 128, 0], [128, 128, 0],
    [0, 0, 128], [128, 0, 128], [0, 128, 128], [192, 192, 192],
    [128, 128, 128], [255, 0, 0], [0, 255, 0], [255, 255, 0],
    [0, 0, 255], [255, 0, 255], [0, 255, 255], [255, 255, 255],
  ];
  for (let i = 0; i < 16; i++) {
    palette[i] = base16[i];
  }

  const cube = [0, 95, 135, 175, 215, 255];
  let idx = 16;
  for (let r = 0; r < 6; r++) {
    for (let g = 0; g < 6; g++) {
      for (let b = 0; b < 6; b++) {
        palette[idx++] = [cube[r], cube[g], cube[b]];
      }
    }
  }

  for (let i = 232; i < 256; i++) {
    const gray = 8 + (i - 232) * 10;
    palette[i] = [gray, gray, gray];
  }

  return palette;
})();

function toHex(n) {
  return n.toString(16).padStart(2, '0');
}

function colorHex(colorCode) {
  const [r, g, b] = ansi256[colorCode];
  return toHex(r) + toHex(g) + toHex(b);
}

// Pre-computed HTML style fragments for all 256 colors
const htmlFgStyle = new Array(256);
const htmlBgStyle = new Array(256);
for (let i = 0; i < 256; i++) {
  const hex = colorHex(i);
  htmlFgStyle[i] = 'color:#' + hex + ';';
  htmlBgStyle[i] = 'background-color:#' + hex + ';';
}

const DEFAULT_COLOR = -2;
const HTML_RESET_ALL = '</span>';

const defaultColorAliases = {
  'black': 0, 'red': 1, 'green': 2, 'yellow': 3,
  'blue': 4, 'magenta': 5, 'cyan': 6, 'white': 7,
  'black-bold': 8, 'red-bold': 9, 'green-bold': 10, 'yellow-bold': 11,
  'blue-bold': 12, 'magenta-bold': 13, 'cyan-bold': 14, 'white-bold': 15,
};

const defaultPositionMap = {
  'topleft': ['1', '1'],
};

const clearMap = {
  'aftercursor': 0,
  'beforecursor': 1,
  'all': 2,
  'scrollback': 3,
};

// Mutable state — aliases and position map
let colorAliases = Object.assign({}, defaultColorAliases);
let positionMap = Object.assign({}, defaultPositionMap);

// --- Tag matcher ---

class TagMatcher {
  constructor(startByte, midBytes, endByte, unknownLength) {
    this.startByte = startByte;
    this.midBytes = midBytes;
    this.endByte = endByte;
    this.exactMatch = !unknownLength;
    this.totalSize = midBytes.length + 1;
    this.position = 0;
  }

  reset() {
    this.position = 0;
  }

  // Returns { matched: bool, complete: bool }
  matchNext(char) {
    if (this.position === 0) {
      if (char === this.startByte) {
        this.position++;
        return { matched: true, complete: false };
      }
      return { matched: false, complete: true };
    }

    if (this.position >= this.totalSize) {
      if (char === this.endByte) {
        this.position++;
        return { matched: true, complete: true };
      }
      if (this.exactMatch) {
        this.position = 0;
        return { matched: false, complete: true };
      }
      return { matched: true, complete: false };
    }

    if (char === this.midBytes[this.position - 1]) {
      this.position++;
      return { matched: true, complete: false };
    }

    this.position = 0;
    return { matched: false, complete: true };
  }
}

// --- Property extraction ---

function extractProperties(tagStr) {
  const props = { fg: DEFAULT_COLOR, bg: DEFAULT_COLOR, clear: -1, position: null };

  let i = 0;
  const n = tagStr.length;

  while (i < n) {
    if (tagStr[i] !== ' ') {
      i++;
      continue;
    }
    i++; // consume space

    const keyStart = i;
    while (i < n && tagStr[i] !== '=') i++;
    if (i >= n) break;
    const key = tagStr.slice(keyStart, i);
    i++; // consume '='

    if (i >= n) break;

    let quote = 0;
    if (tagStr[i] === "'" || tagStr[i] === '"') {
      quote = tagStr[i];
      i++;
    }
    const valStart = i;
    if (quote) {
      while (i < n && tagStr[i] !== quote) i++;
    } else {
      while (i < n && tagStr[i] !== ' ' && tagStr[i] !== '>') i++;
    }
    const val = tagStr.slice(valStart, i);
    if (quote && i < n) i++; // consume closing quote

    if (val.length === 0) continue;

    switch (key) {
      case 'fg': {
        const num = parseInt(val, 10);
        if (!isNaN(num) && String(num) === val) {
          props.fg = num;
        } else if (colorAliases[val] !== undefined) {
          props.fg = colorAliases[val];
        } else {
          props.fg = DEFAULT_COLOR;
        }
        break;
      }
      case 'bg': {
        const num = parseInt(val, 10);
        if (!isNaN(num) && String(num) === val) {
          props.bg = num;
        } else if (colorAliases[val] !== undefined) {
          props.bg = colorAliases[val];
        } else {
          props.bg = DEFAULT_COLOR;
        }
        break;
      }
      case 'position': {
        let posArr = null;
        if (positionMap[val]) {
          posArr = positionMap[val];
        } else {
          const comma = val.indexOf(',');
          if (comma > 0) {
            posArr = [val.slice(0, comma), val.slice(comma + 1)];
          }
        }
        if (posArr && posArr.length === 2) {
          const x = parseInt(posArr[0], 10);
          const y = parseInt(posArr[1], 10);
          if (!isNaN(x) && !isNaN(y) && x >= 0 && y >= 0 && x <= 16000 && y <= 16000) {
            props.position = [x, y];
          }
        }
        break;
      }
      case 'clear': {
        if (clearMap[val] !== undefined) {
          props.clear = clearMap[val];
        }
        break;
      }
    }
  }

  return props;
}

// --- HTML span generation ---

function propagateHTML(props, previous) {
  let fg = props.fg;
  let bg = props.bg;

  if (previous !== null) {
    if (fg === DEFAULT_COLOR) fg = previous.fg;
    if (bg === DEFAULT_COLOR) bg = previous.bg;
  }

  if (previous !== null && fg === previous.fg && bg === previous.bg) {
    return '<span>';
  }

  if (fg === DEFAULT_COLOR && bg === DEFAULT_COLOR) {
    return '<span>';
  }

  if (fg > -1 && bg > -1) {
    return '<span style="' + htmlFgStyle[fg] + htmlBgStyle[bg] + '">';
  }
  if (fg > -1) {
    return '<span style="' + htmlFgStyle[fg] + '">';
  }
  if (bg > -1) {
    return '<span style="' + htmlBgStyle[bg] + '">';
  }
  return '<span>';
}

// --- Core parser ---

const TAG_START = '<';
const TAG_END = '>';
const TAG_OPEN_MID = 'ansi';
const TAG_CLOSE_MID = '/ansi';
const MAX_TAG_SIZE = 256;

const PARSE_MODE_NONE = 0;
const PARSE_MODE_MATCHING = 1;

/**
 * Parse a string containing <ansi> tags and return HTML with <span> tags.
 *
 * @param {string} str - Input string with ansi tags.
 * @param {object} [options]
 * @param {boolean} [options.stripTags=false] - Remove all valid ansi tags from output.
 * @param {boolean} [options.monochrome=false] - Ignore color properties.
 * @returns {string} Parsed output string.
 */
function parse(str, options) {
  const stripAllTags = !!(options && options.stripTags);
  const stripAllColor = !!(options && options.monochrome);

  const tagStack = [];
  const tagBuf = new Array(MAX_TAG_SIZE);
  let tagLen = 0;

  const openMatcher = new TagMatcher(TAG_START, TAG_OPEN_MID, TAG_END, true);
  const closeMatcher = new TagMatcher(TAG_START, TAG_CLOSE_MID, TAG_END, false);

  let mode = PARSE_MODE_NONE;
  let out = '';

  for (let i = 0; i < str.length; i++) {
    const input = str[i];

    if (mode === PARSE_MODE_NONE) {
      if (input !== TAG_START) {
        out += input;
        continue;
      }
      mode = PARSE_MODE_MATCHING;
    }

    if (mode === PARSE_MODE_MATCHING) {
      const openResult = openMatcher.matchNext(input);
      const closeResult = closeMatcher.matchNext(input);

      if (openResult.matched) {
        if (tagLen < MAX_TAG_SIZE) {
          tagBuf[tagLen++] = input;
        }

        if (!openResult.complete) continue;

        let newTag = extractProperties(tagBuf.slice(0, tagLen).join(''));

        if (stripAllColor) {
          newTag.fg = DEFAULT_COLOR;
          newTag.bg = DEFAULT_COLOR;
        }

        tagLen = 0;

        if (!stripAllTags) {
          const stackLen = tagStack.length;
          const previous = stackLen > 0 ? tagStack[stackLen - 1] : null;
          out += propagateHTML(newTag, previous);
          tagStack.push(newTag);
        }

        mode = PARSE_MODE_NONE;
        openMatcher.reset();
        closeMatcher.reset();
        continue;
      }
      openMatcher.reset();

      if (closeResult.matched) {
        if (tagLen < MAX_TAG_SIZE) {
          tagBuf[tagLen++] = input;
        }

        if (!closeResult.complete) continue;

        tagLen = 0;
        if (!stripAllTags) {
          const stackLen = tagStack.length;

          if (stackLen > 2) {
            out += propagateHTML(tagStack[stackLen - 2], tagStack[stackLen - 3]);
          } else if (stackLen > 1) {
            out += propagateHTML(tagStack[stackLen - 2], null);
          } else {
            out += HTML_RESET_ALL;
          }

          if (stackLen > 0) {
            tagStack.pop();
          }
        }

        mode = PARSE_MODE_NONE;
        openMatcher.reset();
        closeMatcher.reset();
        continue;
      }
      closeMatcher.reset();

      if (openResult.complete && closeResult.complete) {
        if (tagLen < MAX_TAG_SIZE) {
          tagBuf[tagLen++] = input;
        }
      }

      mode = PARSE_MODE_NONE;

      if (!stripAllTags) {
        out += tagBuf.slice(0, tagLen).join('');
      }
      tagLen = 0;
      continue;
    }
  }

  if (!stripAllTags) {
    if (tagLen > 0) {
      out += tagBuf.slice(0, tagLen).join('');
      tagLen = 0;
    }

    if (tagStack.length > 0) {
      out += HTML_RESET_ALL;
    }
  }

  return out;
}

/**
 * Set a single color alias.
 *
 * @param {string} alias
 * @param {number} value - 0-255
 */
function setAlias(alias, value) {
  if (value < 0 || value > 255) {
    throw new RangeError(`value "${value}" out of allowable range for alias "${alias}"`);
  }
  colorAliases = Object.assign({}, colorAliases, { [alias]: value });
}

/**
 * Set multiple color aliases at once.
 *
 * @param {Object.<string, number>} aliases
 */
function setAliases(aliases) {
  for (const [alias, value] of Object.entries(aliases)) {
    if (value < 0 || value > 255) {
      throw new RangeError(`value "${value}" out of allowable range for alias "${alias}"`);
    }
  }
  colorAliases = Object.assign({}, colorAliases, aliases);
}

/**
 * Load color aliases from a flat object.
 *
 * Values may be numeric color codes (0-255) or strings referencing another
 * alias name. Alias-to-alias references are resolved after all numeric values
 * are registered. Unresolvable references are silently ignored.
 *
 * Example:
 *   loadAliases({ date: 207, username: 195, highlight: 'yellow' });
 *
 * @param {Object.<string, number|string>} aliases
 */
function loadAliases(aliases) {
  const newColors = Object.assign({}, colorAliases);
  const deferred = {};

  for (const [alias, raw] of Object.entries(aliases)) {
    const num = typeof raw === 'number' ? raw : parseInt(raw, 10);
    if (!isNaN(num) && String(num) === String(raw).trim()) {
      if (num < 0 || num > 255) {
        throw new RangeError(`value "${num}" out of allowable range for alias "${alias}"`);
      }
      newColors[alias] = num;
    } else {
      deferred[alias] = String(raw);
    }
  }

  for (const [alias, ref] of Object.entries(deferred)) {
    if (newColors[ref] !== undefined) {
      newColors[alias] = newColors[ref];
    }
  }

  colorAliases = newColors;
}

/**
 * Returns the RGB components for an ANSI 256 color code.
 *
 * @param {number} colorCode - 0-255
 * @returns {{ r: number, g: number, b: number, hex: string }}
 */
function rgb(colorCode) {
  if (colorCode < 0 || colorCode > 255) {
    return { r: 0, g: 0, b: 0, hex: '000000' };
  }
  const [r, g, b] = ansi256[colorCode];
  return { r, g, b, hex: colorHex(colorCode) };
}

const _exports = { parse, setAlias, setAliases, loadAliases, rgb };

if (typeof module !== 'undefined' && module.exports) {
  module.exports = _exports;
} else if (typeof window !== 'undefined') {
  window.ansitags = _exports;
}
