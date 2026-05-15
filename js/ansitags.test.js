'use strict';

const { parse, splitString, splitStringOnSpaces, setAlias, setAliases, loadAliases, rgb } = require('./ansitags');

function test(name, fn) {
  try {
    fn();
    console.log('  PASS:', name);
  } catch (e) {
    console.error('  FAIL:', name);
    console.error('       ', e.message);
    process.exitCode = 1;
  }
}

function assertEqual(actual, expected, msg) {
  if (actual !== expected) {
    throw new Error(
      (msg ? msg + '\n' : '') +
      '  expected: ' + JSON.stringify(expected) + '\n' +
      '  actual:   ' + JSON.stringify(actual)
    );
  }
}

// --- HTML mode tests (mirrors testdata/ansitags_test_html.yaml) ---

console.log('\nHTML mode:');

test('No Tag', () => {
  assertEqual(
    parse('This string has no ansi tags'),
    'This string has no ansi tags'
  );
});

test('Named Colors', () => {
  assertEqual(
    parse('<ansi fg=black bg="white">A</ansi><ansi fg="red" bg="cyan">B</ansi><ansi fg="green" bg="magenta">C</ansi><ansi fg="yellow" bg="blue">D</ansi><ansi fg="blue" bg="yellow">E</ansi><ansi fg="magenta" bg="green">F</ansi><ansi fg="cyan" bg="red">G</ansi><ansi fg="white" bg="black">H</ansi>'),
    '<span style="color:#000000;background-color:#c0c0c0;">A</span><span style="color:#800000;background-color:#008080;">B</span><span style="color:#008000;background-color:#800080;">C</span><span style="color:#808000;background-color:#000080;">D</span><span style="color:#000080;background-color:#808000;">E</span><span style="color:#800080;background-color:#008000;">F</span><span style="color:#008080;background-color:#800000;">G</span><span style="color:#c0c0c0;background-color:#000000;">H</span>'
  );
});

test('Single Tag', () => {
  assertEqual(
    parse("<ansi fg='blue'>This is inside of ansi tags</ansi>"),
    '<span style="color:#000080;">This is inside of ansi tags</span>'
  );
});

test('Nested Tag', () => {
  assertEqual(
    parse('<ansi fg="blue" bg="green">This is <ansi fg="blue" bg="green" bold="true">inside</ansi> of ansi tags</ansi>'),
    '<span style="color:#000080;background-color:#008000;">This is <span>inside<span style="color:#000080;background-color:#008000;"> of ansi tags</span>'
  );
});

test('Single Tag IN normal text', () => {
  assertEqual(
    parse('Prefix text <ansi fg="blue" bg="black" bold="true">This is inside of ansi tags</ansi> suffix text'),
    'Prefix text <span style="color:#000080;background-color:#000000;">This is inside of ansi tags</span> suffix text'
  );
});

test('Many Nested Tags', () => {
  assertEqual(
    parse('[one]<ansi fg="green" bg="blue" >[two]<ansi fg="black" bg="yellow" >[t<ansi fg="green" bg="magenta">h<ansi fg="black" bg="red">r</ansi>e</ansi>e]</ansi>[four]</ansi>[five]'),
    '[one]<span style="color:#008000;background-color:#000080;">[two]<span style="color:#000000;background-color:#808000;">[t<span style="color:#008000;background-color:#800080;">h<span style="color:#000000;background-color:#800000;">r<span style="color:#008000;background-color:#800080;">e<span style="color:#000000;background-color:#808000;">e]<span style="color:#008000;background-color:#000080;">[four]</span>[five]'
  );
});

test('Multiple sequential Tags', () => {
  assertEqual(
    parse('start normal text <ansi fg="blue" bg="yellow">tagged text 1</ansi> <ansi fg="blue" bg="107">tagged text 2</ansi> <ansi fg="blue" bg="yellow" bold="true">tagged text 3</ansi> end normal text'),
    'start normal text <span style="color:#000080;background-color:#808000;">tagged text 1</span> <span style="color:#000080;background-color:#87af5f;">tagged text 2</span> <span style="color:#000080;background-color:#808000;">tagged text 3</span> end normal text'
  );
});

test('No Close Tag', () => {
  assertEqual(
    parse("<ansi fg='blue'>This is inside of ansi tags"),
    '<span style="color:#000080;">This is inside of ansi tags</span>'
  );
});

test('No Tags', () => {
  assertEqual(
    parse('This is inside of ansi tags'),
    'This is inside of ansi tags'
  );
});

test('Unterminated Open Tag', () => {
  assertEqual(
    parse("<ansi fg='blue'This is inside of ansi tags</ansi>"),
    '<span style="color:#000080;"></span>'
  );
});

test('Unterminated Open Tag and No Close Tag', () => {
  assertEqual(
    parse("<ansi fg='blue'This is inside of ansi tags"),
    "<ansi fg='blue'This is inside of ansi tags"
  );
});

test('Leading Close Tag', () => {
  assertEqual(
    parse("</ansi><ansi fg='blue'>This is inside of ansi tags</ansi>"),
    '</span><span style="color:#000080;">This is inside of ansi tags</span>'
  );
});

test('Crossed malformed Tags', () => {
  assertEqual(
    parse("<ansi fg='blue' </ansi >This is inside of ansi tags>"),
    '<span style="color:#000080;">This is inside of ansi tags></span>'
  );
});

test('Empty tags', () => {
  assertEqual(
    parse('<ansi>This is inside of ansi tags</ansi>'),
    '<span>This is inside of ansi tags</span>'
  );
});

test('Background only', () => {
  assertEqual(
    parse('<ansi bg="magenta">This is inside of ansi tags</ansi>'),
    '<span style="background-color:#800080;">This is inside of ansi tags</span>'
  );
});

test('Invalid color strings', () => {
  assertEqual(
    parse('<ansi fg="tomato" bg="fish">This is inside of ansi tags</ansi>'),
    '<span>This is inside of ansi tags</span>'
  );
});

test('Nesting', () => {
  assertEqual(
    parse('<ansi fg="10">.:</ansi> <ansi fg="226"><ansi fg="196">Test</ansi> More</ansi>'),
    '<span style="color:#00ff00;">.:</span> <span style="color:#ffff00;"><span style="color:#ff0000;">Test<span style="color:#ffff00;"> More</span>'
  );
});

test('Nested unrecognized alias', () => {
  assertEqual(
    parse('<ansi fg="8">For those about to <ansi fg="unrecognizedalias">Rock</ansi> we <ansi fg="15">salute</ansi> you.</ansi>'),
    '<span style="color:#808080;">For those about to <span>Rock<span style="color:#808080;"> we <span style="color:#ffffff;">salute<span style="color:#808080;"> you.</span>'
  );
});

// --- Strip mode tests ---

console.log('\nStrip mode:');

test('No Tag', () => {
  assertEqual(
    parse('This string has no ansi tags', { stripTags: true }),
    'This string has no ansi tags'
  );
});

test('Named Colors', () => {
  assertEqual(
    parse('<ansi fg=black bg="white">A</ansi><ansi fg="red" bg="cyan">B</ansi><ansi fg="green" bg="magenta">C</ansi><ansi fg="yellow" bg="blue">D</ansi><ansi fg="blue" bg="yellow">E</ansi><ansi fg="magenta" bg="green">F</ansi><ansi fg="cyan" bg="red">G</ansi><ansi fg="white" bg="black">H</ansi>', { stripTags: true }),
    'ABCDEFGH'
  );
});

test('No Close Tag', () => {
  assertEqual(
    parse("<ansi fg='blue'>This is inside of ansi tags", { stripTags: true }),
    'This is inside of ansi tags'
  );
});

// --- Monochrome mode tests ---

console.log('\nMonochrome mode (HTML):');

test('Named Colors monochrome', () => {
  const result = parse('<ansi fg=black bg="white">A</ansi>', { monochrome: true });
  assertEqual(result, '<span>A</span>');
});

// --- setAlias / setAliases tests ---

console.log('\nAlias tests:');

test('setAlias', () => {
  setAlias('testcolor', 207);
  assertEqual(
    parse("<ansi fg='testcolor'>hello</ansi>"),
    '<span style="color:#ff5fff;">hello</span>'
  );
});

test('setAliases', () => {
  setAliases({ 'myred': 196, 'myblue': 21 });
  assertEqual(
    parse('<ansi fg="myred" bg="myblue">hi</ansi>'),
    '<span style="color:#ff0000;background-color:#0000ff;">hi</span>'
  );
});

test('setAlias out of range throws', () => {
  let threw = false;
  try { setAlias('bad', 256); } catch (e) { threw = true; }
  if (!threw) throw new Error('expected RangeError');
});

// --- loadAliases tests ---

console.log('\nloadAliases tests:');

test('numeric values', () => {
  loadAliases({ date: 207, username: 195 });
  assertEqual(
    parse('<ansi fg="date">hello</ansi>'),
    '<span style="color:#ff5fff;">hello</span>'
  );
  assertEqual(
    parse('<ansi fg="username">world</ansi>'),
    '<span style="color:#d7ffff;">world</span>'
  );
});

test('alias-to-alias reference', () => {
  loadAliases({ mygreen: 'green' });
  assertEqual(
    parse('<ansi fg="mygreen">hi</ansi>'),
    '<span style="color:#008000;">hi</span>'
  );
});

test('color256 group', () => {
  loadAliases({ specialred: 196 });
  assertEqual(
    parse('<ansi fg="specialred">hi</ansi>'),
    '<span style="color:#ff0000;">hi</span>'
  );
});

test('out of range throws', () => {
  let threw = false;
  try { loadAliases({ bad: 300 }); } catch (e) { threw = true; }
  if (!threw) throw new Error('expected RangeError');
});

test('unresolvable alias-to-alias is silently ignored', () => {
  loadAliases({ ghost: 'nonexistent' });
  assertEqual(
    parse('<ansi fg="ghost">hi</ansi>'),
    '<span>hi</span>'
  );
});

// --- rgb() tests ---

console.log('\nRGB tests:');

test('rgb(0) is black', () => {
  const c = rgb(0);
  assertEqual(c.r, 0); assertEqual(c.g, 0); assertEqual(c.b, 0);
  assertEqual(c.hex, '000000');
});

test('rgb(15) is white', () => {
  const c = rgb(15);
  assertEqual(c.r, 255); assertEqual(c.g, 255); assertEqual(c.b, 255);
  assertEqual(c.hex, 'ffffff');
});

test('rgb(201) is magenta', () => {
  const c = rgb(201);
  assertEqual(c.hex, 'ff00ff');
});

test('rgb out of range returns black', () => {
  const c = rgb(300);
  assertEqual(c.hex, '000000');
});

// --- splitString tests ---

console.log('\nsplitString tests:');

function assertArrayEqual(actual, expected, msg) {
  if (!Array.isArray(actual) || !Array.isArray(expected)) {
    throw new Error((msg ? msg + ': ' : '') + 'expected arrays');
  }
  if (actual.length !== expected.length) {
    throw new Error(
      (msg ? msg + '\n' : '') +
      '  length mismatch: expected ' + expected.length + ', got ' + actual.length + '\n' +
      '  expected: ' + JSON.stringify(expected) + '\n' +
      '  actual:   ' + JSON.stringify(actual)
    );
  }
  for (let i = 0; i < expected.length; i++) {
    if (actual[i] !== expected[i]) {
      throw new Error(
        (msg ? msg + '\n' : '') +
        '  index ' + i + ':\n' +
        '  expected: ' + JSON.stringify(expected[i]) + '\n' +
        '  actual:   ' + JSON.stringify(actual[i])
      );
    }
  }
}

function strippedConcat(segments) {
  return segments.map(s => parse(s, { stripTags: true })).join('');
}

function segVisibleLen(seg) {
  return parse(seg, { stripTags: true }).length;
}

test('no tags basic split', () => {
  assertArrayEqual(splitString('Hello World', 5, false), ['Hello', ' Worl', 'd']);
});

test('no split needed', () => {
  const input = '<ansi fg="red">Hi</ansi>';
  assertArrayEqual(splitString(input, 10, false), [input]);
});

test('exact length no split', () => {
  const input = '<ansi fg="red">Hello</ansi>';
  assertArrayEqual(splitString(input, 5, false), [input]);
});

test('empty input and zero maxLen', () => {
  assertArrayEqual(splitString('', 10, false), ['']);
  assertArrayEqual(splitString('hello', 0, false), ['hello']);
});

test('nested tags split', () => {
  const input = '<ansi fg="yellow">This is some <ansi fg="black">long as heck</ansi> text</ansi>';
  assertArrayEqual(splitString(input, 17, false), [
    '<ansi fg="yellow">This is some <ansi fg="black">long</ansi></ansi>',
    '<ansi fg="yellow"><ansi fg="black"> as heck</ansi> text</ansi>',
  ]);
});

test('multiple splits single tag', () => {
  assertArrayEqual(splitString('<ansi fg="red">abcdefghijklmno</ansi>', 5, false), [
    '<ansi fg="red">abcde</ansi>',
    '<ansi fg="red">fghij</ansi>',
    '<ansi fg="red">klmno</ansi>',
  ]);
});

test('maxLen 1 per char', () => {
  const result = splitString('<ansi fg="red">ABCDE</ansi>', 1, false);
  assertEqual(result.length, 5);
  for (let i = 0; i < 5; i++) {
    assertEqual(result[i], '<ansi fg="red">' + 'ABCDE'[i] + '</ansi>');
  }
});

test('tag closes exactly at split', () => {
  const result = splitString('<ansi fg="red">12345</ansi>67890ABCDE', 5, false);
  assertEqual(result.length, 3);
  assertEqual(result[0], '<ansi fg="red">12345</ansi>');
  assertEqual(result[1], '<ansi fg="red"></ansi>67890');
  assertEqual(result[2], 'ABCDE');
});

test('preserves all chars across splits', () => {
  const input = '<ansi fg="red">Hello <ansi fg="green">World, this is a <ansi fg="blue">deeply nested and very long</ansi> string that</ansi> continues here</ansi>';
  const result = splitString(input, 15, false);

  assertEqual(strippedConcat(result), parse(input, { stripTags: true }));
  for (let i = 0; i < result.length - 1; i++) {
    assertEqual(segVisibleLen(result[i]), 15, 'segment ' + i);
  }
});

test('all segments parseable for various inputs and lengths', () => {
  const inputs = [
    '<ansi fg="red">Simple</ansi>',
    '<ansi fg="red"><ansi fg="green"><ansi fg="blue">Triple nested long text here</ansi></ansi></ansi>',
    'Before<ansi fg="red">middle</ansi>after',
    '<ansi fg="red" bg="blue">Mixed <ansi fg="green">attributes <ansi fg="yellow" bg="white">everywhere</ansi> in this</ansi> string</ansi>',
  ];

  for (const input of inputs) {
    for (let maxLen = 1; maxLen <= 10; maxLen++) {
      const result = splitString(input, maxLen, false);
      for (let i = 0; i < result.length; i++) {
        parse(result[i]);
        const vl = segVisibleLen(result[i]);
        if (vl > maxLen) throw new Error('input=' + JSON.stringify(input) + ' maxLen=' + maxLen + ' seg=' + i + ' visible=' + vl);
      }
      const original = parse(input, { stripTags: true });
      assertEqual(strippedConcat(result), original, 'input=' + JSON.stringify(input) + ' maxLen=' + maxLen);
    }
  }
});

// --- splitString trimSpace tests ---

console.log('\nsplitString trimSpace tests:');

test('trim default', () => {
  const input = '<ansi fg="yellow">This is some <ansi fg="black">long as heck</ansi> text</ansi>';
  const result = splitString(input, 17);
  assertEqual(result.length, 2);
  assertEqual(result[0], '<ansi fg="yellow">This is some <ansi fg="black">long</ansi></ansi>');
  assertEqual(result[1], '<ansi fg="yellow"><ansi fg="black">as heck</ansi> text</ansi>');
});

test('trim leading and trailing spaces', () => {
  assertArrayEqual(splitString('Hello World', 5), ['Hello', 'Worl', 'd']);

  const result = splitString('<ansi fg="red">Hello World</ansi>', 5);
  assertEqual(result[0], '<ansi fg="red">Hello</ansi>');
  assertEqual(result[1], '<ansi fg="red">Worl</ansi>');
});

test('trim preserves internal spaces', () => {
  const result = splitString('<ansi fg="red">a b c d e f g</ansi>', 5);
  assertEqual(result.length, 3);
  assertEqual(result[0], '<ansi fg="red">a b c</ansi>');
  assertEqual(result[1], '<ansi fg="red">d e</ansi>');
  assertEqual(result[2], '<ansi fg="red">f g</ansi>');
});

test('trim false preserves spaces', () => {
  const result = splitString('<ansi fg="red">Hello World</ansi>', 5, false);
  assertEqual(result.length, 3);
  assertEqual(result[0], '<ansi fg="red">Hello</ansi>');
  assertEqual(result[1], '<ansi fg="red"> Worl</ansi>');
  assertEqual(result[2], '<ansi fg="red">d</ansi>');
});

// --- splitStringOnSpaces tests ---

console.log('\nsplitStringOnSpaces tests:');

test('basic word split', () => {
  assertArrayEqual(splitStringOnSpaces('Hello World', 8), ['Hello', 'World']);
});

test('no split needed', () => {
  const input = '<ansi fg="red">Hello</ansi>';
  assertArrayEqual(splitStringOnSpaces(input, 10), [input]);
});

test('empty input and zero maxLen', () => {
  assertArrayEqual(splitStringOnSpaces('', 10), ['']);
  assertArrayEqual(splitStringOnSpaces('hello', 0), ['hello']);
});

test('fallback to char split when no space', () => {
  assertArrayEqual(splitStringOnSpaces('ABCDEFGHIJ', 4, false), ['ABCD', 'EFGH', 'IJ']);
});

test('space at boundary causes char split then space split', () => {
  // "Hello World" maxLen=5, trimSpace=false:
  // char split at 5, then space split at 6 -> ["Hello", " ", "World"]
  const result = splitStringOnSpaces('Hello World', 5, false);
  assertEqual(result.length, 3);
  assertEqual(result[0], 'Hello');
  assertEqual(result[1], ' ');
  assertEqual(result[2], 'World');
});

test('splits with tags preserve all chars', () => {
  const input = '<ansi fg="white">The quick <ansi fg="green">brown fox</ansi> jumps over the lazy dog</ansi>';
  const result = splitStringOnSpaces(input, 12, false);

  assertEqual(strippedConcat(result), parse(input, { stripTags: true }));
  for (const seg of result) {
    if (segVisibleLen(seg) > 12) throw new Error('segment exceeds maxLen: ' + JSON.stringify(seg));
  }
});

test('mixed fallback when word exceeds maxLen', () => {
  const result = splitStringOnSpaces('word toolongword next', 8, false);

  assertEqual(strippedConcat(result), 'word toolongword next');
  for (const seg of result) {
    if (segVisibleLen(seg) > 8) throw new Error('segment exceeds maxLen: ' + JSON.stringify(seg));
  }
});

test('splits at space within tagged text', () => {
  const input = '<ansi fg="yellow">Hello <ansi fg="red">World foo</ansi> bar</ansi>';
  const result = splitStringOnSpaces(input, 10);

  assertEqual(result.length, 3);
  assertEqual(parse(result[0], { stripTags: true }), 'Hello');
  assertEqual(parse(result[1], { stripTags: true }), 'World foo');
  assertEqual(parse(result[2], { stripTags: true }), 'bar');
});

console.log('');
