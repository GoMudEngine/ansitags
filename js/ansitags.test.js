'use strict';

const { parse, splitString, setAlias, setAliases, loadAliases, rgb } = require('./ansitags');

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

test('nested tags basic', () => {
  const input = '<ansi fg="yellow">This is some <ansi fg="black">long as heck</ansi> text</ansi>';
  const result = splitString(input, 17, false);
  assertArrayEqual(result, [
    '<ansi fg="yellow">This is some <ansi fg="black">long</ansi></ansi>',
    '<ansi fg="yellow"><ansi fg="black"> as heck</ansi> text</ansi>',
  ]);
});

test('no tags plain text', () => {
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

test('empty input', () => {
  assertArrayEqual(splitString('', 10, false), ['']);
});

test('zero maxLen', () => {
  assertArrayEqual(splitString('hello', 0, false), ['hello']);
});

test('multiple splits single tag', () => {
  const input = '<ansi fg="red">abcdefghijklmno</ansi>';
  const result = splitString(input, 5, false);
  assertArrayEqual(result, [
    '<ansi fg="red">abcde</ansi>',
    '<ansi fg="red">fghij</ansi>',
    '<ansi fg="red">klmno</ansi>',
  ]);
});

test('tag at split boundary', () => {
  const result = splitString('AB<ansi fg="red">CD</ansi>EF', 3, false);
  assertArrayEqual(result, [
    'AB<ansi fg="red">C</ansi>',
    '<ansi fg="red">D</ansi>EF',
  ]);
});

test('deeply nested 3 levels', () => {
  const input = '<ansi fg="red"><ansi fg="green"><ansi fg="blue">Hello World</ansi></ansi></ansi>';
  const result = splitString(input, 5, false);
  assertEqual(result.length, 3);
  assertEqual(result[0], '<ansi fg="red"><ansi fg="green"><ansi fg="blue">Hello</ansi></ansi></ansi>');
  assertEqual(result[1], '<ansi fg="red"><ansi fg="green"><ansi fg="blue"> Worl</ansi></ansi></ansi>');
  assertEqual(result[2], '<ansi fg="red"><ansi fg="green"><ansi fg="blue">d</ansi></ansi></ansi>');
});

test('tag closes before split', () => {
  const result = splitString('<ansi fg="red">AB</ansi>CDEF', 3, false);
  assertEqual(result.length, 2);
  assertEqual(result[0], '<ansi fg="red">AB</ansi>C');
  assertEqual(result[1], 'DEF');
});

test('maxLen 1 per char', () => {
  const result = splitString('<ansi fg="red">ABCDE</ansi>', 1, false);
  assertEqual(result.length, 5);
  for (let i = 0; i < 5; i++) {
    assertEqual(result[i], '<ansi fg="red">' + 'ABCDE'[i] + '</ansi>');
  }
});

test('long multi-color paragraph', () => {
  const input =
    '<ansi fg="red">The quick </ansi>' +
    '<ansi fg="green">brown fox </ansi>' +
    '<ansi fg="blue">jumps over </ansi>' +
    '<ansi fg="yellow">the lazy </ansi>' +
    '<ansi fg="magenta">dog and then runs away</ansi>';
  const result = splitString(input, 20, false);

  assertEqual(result.length, 4);
  assertEqual(segVisibleLen(result[0]), 20);
  assertEqual(segVisibleLen(result[1]), 20);
  assertEqual(segVisibleLen(result[2]), 20);
  assertEqual(segVisibleLen(result[3]), 2);

  const original = parse(input, { stripTags: true });
  assertEqual(strippedConcat(result), original);
});

test('nested tags open and close across multiple splits', () => {
  const input = '<ansi fg="red">Hello <ansi fg="green">World, this is a <ansi fg="blue">deeply nested and very long</ansi> string that</ansi> continues outside the inner tags with more text here</ansi>';
  const result = splitString(input, 15, false);

  const expectedVisible = 'Hello World, this is a deeply nested and very long string that continues outside the inner tags with more text here';
  assertEqual(parse(input, { stripTags: true }).length, expectedVisible.length);

  for (let i = 0; i < result.length - 1; i++) {
    assertEqual(segVisibleLen(result[i]), 15, 'segment ' + i);
  }
  assertEqual(strippedConcat(result), expectedVisible);
});

test('many sequential tags', () => {
  const input =
    '<ansi fg="red">AB</ansi><ansi fg="green">CD</ansi><ansi fg="blue">EF</ansi>' +
    '<ansi fg="yellow">GH</ansi><ansi fg="magenta">IJ</ansi><ansi fg="cyan">KL</ansi>' +
    '<ansi fg="white">MN</ansi><ansi fg="red">OP</ansi>';
  const result = splitString(input, 3, false);

  assertEqual(result.length, 6);
  assertEqual(strippedConcat(result), 'ABCDEFGHIJKLMNOP');
  assertEqual(result[0], '<ansi fg="red">AB</ansi><ansi fg="green">C</ansi>');
  assertEqual(result[1], '<ansi fg="green">D</ansi><ansi fg="blue">EF</ansi>');
});

test('alternating tagged and untagged', () => {
  const input = 'plain1<ansi fg="red">RED</ansi>plain2<ansi fg="blue">BLUE</ansi>plain3<ansi fg="green">GREEN</ansi>plain4';
  const result = splitString(input, 10, false);

  assertEqual(result.length, 4);
  assertEqual(strippedConcat(result), 'plain1REDplain2BLUEplain3GREENplain4');
  for (let i = 0; i < result.length - 1; i++) {
    assertEqual(segVisibleLen(result[i]), 10, 'segment ' + i);
  }
});

test('deep 5-level nesting', () => {
  const input =
    '<ansi fg="red"><ansi fg="green"><ansi fg="blue"><ansi fg="yellow"><ansi fg="magenta">' +
    'This text is five levels deep and should be split properly across segments' +
    '</ansi></ansi></ansi></ansi></ansi>';
  const result = splitString(input, 12, false);

  const expectedVisible = 'This text is five levels deep and should be split properly across segments';
  assertEqual(result.length, 7);
  assertEqual(strippedConcat(result), expectedVisible);

  for (let i = 0; i < result.length - 1; i++) {
    assertEqual(segVisibleLen(result[i]), 12, 'segment ' + i);
  }

  // Middle segments must reopen all 5 levels
  for (let i = 1; i < result.length - 1; i++) {
    if (result[i].indexOf('<ansi fg="red">') === -1) throw new Error('segment ' + i + ' missing red');
    if (result[i].indexOf('<ansi fg="magenta">') === -1) throw new Error('segment ' + i + ' missing magenta');
  }
});

test('nesting changes across splits', () => {
  const input = '<ansi fg="red">Level one <ansi fg="green">level two here</ansi> back to one</ansi> and now plain';
  const result = splitString(input, 10, false);

  const expectedVisible = 'Level one level two here back to one and now plain';
  assertEqual(result.length, 5);
  assertEqual(strippedConcat(result), expectedVisible);
});

test('with bg attributes', () => {
  const input = '<ansi fg="red" bg="white">Warning: <ansi fg="yellow" bg="black">critical error in module</ansi> please check logs immediately</ansi>';
  const result = splitString(input, 15, false);

  const expectedVisible = 'Warning: critical error in module please check logs immediately';
  assertEqual(strippedConcat(result), expectedVisible);

  if (result[0].indexOf('bg="white"') === -1) throw new Error('segment 0 missing bg="white"');
  if (result.length > 1 && result[1].indexOf('bg="white"') === -1) throw new Error('segment 1 missing bg="white"');
});

test('tag opens exactly at split', () => {
  const result = splitString('12345<ansi fg="red">67890</ansi>', 5, false);
  assertEqual(result.length, 2);
  assertEqual(result[0], '12345');
  assertEqual(result[1], '<ansi fg="red">67890</ansi>');
});

test('tag closes exactly at split', () => {
  const result = splitString('<ansi fg="red">12345</ansi>67890ABCDE', 5, false);
  assertEqual(result.length, 3);
  assertEqual(result[0], '<ansi fg="red">12345</ansi>');
  assertEqual(result[1], '<ansi fg="red"></ansi>67890');
  assertEqual(result[2], 'ABCDE');
});

test('long paragraph multiple tag styles', () => {
  const input =
    '<ansi fg="white">In the beginning, the universe was created. </ansi>' +
    '<ansi fg="yellow">This has made a lot of people very angry </ansi>' +
    '<ansi fg="red">and has been widely regarded as a <ansi fg="white" bg="red">bad move</ansi>. </ansi>' +
    '<ansi fg="green">The ships hung in the sky in much the same way that <ansi fg="cyan">bricks don\'t</ansi>.</ansi>';
  const result = splitString(input, 25, false);

  const expectedVisible = parse(input, { stripTags: true });
  assertEqual(strippedConcat(result), expectedVisible);

  for (let i = 0; i < result.length - 1; i++) {
    assertEqual(segVisibleLen(result[i]), 25, 'segment ' + i);
  }
});

test('repeated split and rejoin for many maxLen values', () => {
  const input =
    '<ansi fg="red">AAA<ansi fg="green">BBB<ansi fg="blue">CCC</ansi>DDD</ansi>EEE</ansi>' +
    '<ansi fg="yellow">FFF<ansi fg="magenta">GGG</ansi>HHH</ansi>' +
    '<ansi fg="cyan">III</ansi>JJJ';

  const expectedVisible = parse(input, { stripTags: true });

  [1, 2, 3, 4, 5, 7, 10, 13, 15, 29, 30, 100].forEach(maxLen => {
    const result = splitString(input, maxLen, false);
    assertEqual(strippedConcat(result), expectedVisible, 'maxLen=' + maxLen);

    for (let i = 0; i < result.length; i++) {
      const vl = segVisibleLen(result[i]);
      if (vl > maxLen) throw new Error('maxLen=' + maxLen + ' segment ' + i + ' has ' + vl + ' visible chars');
    }

    if (expectedVisible.length > maxLen) {
      for (let i = 0; i < result.length - 1; i++) {
        assertEqual(segVisibleLen(result[i]), maxLen, 'maxLen=' + maxLen + ' segment ' + i);
      }
    }
  });
});

test('single tag spans many segments', () => {
  const input = '<ansi fg="red">abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789</ansi>';
  const result = splitString(input, 8, false);

  assertEqual(result.length, 8);
  assertEqual(strippedConcat(result), 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789');

  for (const seg of result) {
    if (seg.indexOf('<ansi fg="red">') === -1) throw new Error('segment missing open tag');
    if (seg.indexOf('</ansi>') === -1) throw new Error('segment missing close tag');
  }
});

test('nesting depth changes every few chars', () => {
  const input =
    'A<ansi fg="red">B<ansi fg="green">C</ansi>D</ansi>E' +
    '<ansi fg="blue">F<ansi fg="yellow">G<ansi fg="magenta">H</ansi>I</ansi>J</ansi>' +
    'K<ansi fg="cyan">L</ansi>M';
  const result = splitString(input, 4, false);

  assertEqual(strippedConcat(result), 'ABCDEFGHIJKLM');
  for (const seg of result) {
    if (segVisibleLen(seg) > 4) throw new Error('segment exceeds maxLen');
  }
});

test('long real-world example', () => {
  const input =
    '<ansi fg="white">You are standing in a <ansi fg="green">lush forest</ansi>. ' +
    'The trees tower above you, their <ansi fg="green">leaves rustling</ansi> in the wind. ' +
    'A <ansi fg="yellow">narrow path</ansi> leads <ansi fg="red">north</ansi> toward ' +
    'a <ansi fg="magenta">dark cave</ansi>, while another trail winds ' +
    '<ansi fg="cyan">east</ansi> through the <ansi fg="green">underbrush</ansi>. ' +
    '<ansi fg="yellow">A small <ansi fg="red">treasure chest</ansi> sits near the base of an old oak tree</ansi>.</ansi>';
  const result = splitString(input, 40, false);

  const expectedVisible = parse(input, { stripTags: true });
  assertEqual(strippedConcat(result), expectedVisible);

  for (let i = 0; i < result.length - 1; i++) {
    assertEqual(segVisibleLen(result[i]), 40, 'segment ' + i);
  }
});

test('all segments parseable', () => {
  const inputs = [
    '<ansi fg="red">Simple</ansi>',
    '<ansi fg="red"><ansi fg="green"><ansi fg="blue">Triple nested long text here</ansi></ansi></ansi>',
    'Before<ansi fg="red">middle</ansi>after',
    '<ansi fg="red">A</ansi><ansi fg="green">B</ansi><ansi fg="red">C</ansi><ansi fg="green">D</ansi>',
    '<ansi fg="red" bg="blue">Mixed <ansi fg="green">attributes <ansi fg="yellow" bg="white">everywhere</ansi> in this</ansi> string</ansi>',
  ];

  for (const input of inputs) {
    for (let maxLen = 1; maxLen <= 10; maxLen++) {
      const result = splitString(input, maxLen, false);
      for (let i = 0; i < result.length; i++) {
        parse(result[i]);
        parse(result[i], { stripTags: true });
        const vl = segVisibleLen(result[i]);
        if (vl > maxLen) throw new Error('input=' + JSON.stringify(input) + ' maxLen=' + maxLen + ' seg=' + i + ' visible=' + vl);
      }
    }
  }
});

// --- trimSpace tests (default true) ---

console.log('\nsplitString trimSpace tests:');

test('trim default matches user example', () => {
  const input = '<ansi fg="yellow">This is some <ansi fg="black">long as heck</ansi> text</ansi>';
  const result = splitString(input, 17);
  assertEqual(result.length, 2);
  assertEqual(result[0], '<ansi fg="yellow">This is some <ansi fg="black">long</ansi></ansi>');
  assertEqual(result[1], '<ansi fg="yellow"><ansi fg="black">as heck</ansi> text</ansi>');
});

test('trim leading space inside tag', () => {
  const input = '<ansi fg="red">Hello World</ansi>';
  const result = splitString(input, 5);
  assertEqual(result.length, 3);
  assertEqual(result[0], '<ansi fg="red">Hello</ansi>');
  assertEqual(result[1], '<ansi fg="red">Worl</ansi>');
  assertEqual(result[2], '<ansi fg="red">d</ansi>');
});

test('trim trailing space inside tag', () => {
  const input = '<ansi fg="red">test </ansi><ansi fg="blue">more</ansi>';
  const result = splitString(input, 5);
  assertEqual(result.length, 2);
  assertEqual(result[0], '<ansi fg="red">test</ansi>');
  assertEqual(result[1], '<ansi fg="red"></ansi><ansi fg="blue">more</ansi>');
});

test('trim plain text', () => {
  assertArrayEqual(splitString('Hello World', 5), ['Hello', 'Worl', 'd']);
});

test('trim nested tags', () => {
  const input = '<ansi fg="red"><ansi fg="green"><ansi fg="blue">Hello World</ansi></ansi></ansi>';
  const result = splitString(input, 5);
  assertEqual(result.length, 3);
  assertEqual(result[0], '<ansi fg="red"><ansi fg="green"><ansi fg="blue">Hello</ansi></ansi></ansi>');
  assertEqual(result[1], '<ansi fg="red"><ansi fg="green"><ansi fg="blue">Worl</ansi></ansi></ansi>');
  assertEqual(result[2], '<ansi fg="red"><ansi fg="green"><ansi fg="blue">d</ansi></ansi></ansi>');
});

test('trim preserves internal spaces', () => {
  const input = '<ansi fg="red">a b c d e f g</ansi>';
  const result = splitString(input, 5);
  assertEqual(result.length, 3);
  assertEqual(result[0], '<ansi fg="red">a b c</ansi>');
  assertEqual(result[1], '<ansi fg="red">d e</ansi>');
  assertEqual(result[2], '<ansi fg="red">f g</ansi>');
});

test('trim false preserves spaces', () => {
  const input = '<ansi fg="red">Hello World</ansi>';
  const result = splitString(input, 5, false);
  assertEqual(result.length, 3);
  assertEqual(result[0], '<ansi fg="red">Hello</ansi>');
  assertEqual(result[1], '<ansi fg="red"> Worl</ansi>');
  assertEqual(result[2], '<ansi fg="red">d</ansi>');
});

test('trim multiple leading spaces', () => {
  const input = '<ansi fg="red">abc   def</ansi>';
  const result = splitString(input, 4);
  assertEqual(result.length, 3);
  assertEqual(result[0], '<ansi fg="red">abc</ansi>');
  assertEqual(result[1], '<ansi fg="red">de</ansi>');
  assertEqual(result[2], '<ansi fg="red">f</ansi>');
});

test('trim multiple trailing spaces', () => {
  const result = splitString('abc   def', 5);
  assertEqual(result.length, 2);
  assertEqual(result[0], 'abc');
  assertEqual(result[1], 'def');
});

test('trim all-space segment', () => {
  const input = '<ansi fg="red">     hello</ansi>';
  const result = splitString(input, 3);
  assertEqual(result.length, 4);
  assertEqual(result[0], '<ansi fg="red"></ansi>');
  assertEqual(result[1], '<ansi fg="red">h</ansi>');
  assertEqual(result[2], '<ansi fg="red">ell</ansi>');
  assertEqual(result[3], '<ansi fg="red">o</ansi>');
});

test('trim tag close then space', () => {
  const input = '<ansi fg="red">Hello</ansi> <ansi fg="blue">World</ansi>';
  const result = splitString(input, 6);
  assertEqual(result.length, 2);
  assertEqual(result[0], '<ansi fg="red">Hello</ansi>');
  assertEqual(result[1], '<ansi fg="blue">World</ansi>');
});

test('trim long real-world', () => {
  const input =
    '<ansi fg="white">You enter the <ansi fg="green">forest clearing</ansi>. ' +
    'A <ansi fg="yellow">golden light</ansi> shines from above.</ansi>';
  const result = splitString(input, 20);

  for (let i = 0; i < result.length; i++) {
    const stripped = parse(result[i], { stripTags: true });
    if (stripped.length > 0) {
      if (stripped[0] === ' ') throw new Error('segment ' + i + ' has leading space');
      if (stripped[stripped.length - 1] === ' ') throw new Error('segment ' + i + ' has trailing space');
    }
    parse(result[i]);
  }
});

test('trim space between tags', () => {
  const input = '<ansi fg="red">AAAA</ansi> <ansi fg="blue">BBBB</ansi>';
  const result = splitString(input, 4);
  assertEqual(result.length, 3);
  assertEqual(result[0], '<ansi fg="red">AAAA</ansi>');
  assertEqual(result[1], '<ansi fg="red"></ansi><ansi fg="blue">BBB</ansi>');
  assertEqual(result[2], '<ansi fg="blue">B</ansi>');
});

console.log('');
