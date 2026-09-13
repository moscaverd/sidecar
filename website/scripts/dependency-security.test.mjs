import assert from 'node:assert/strict';
import {readFile} from 'node:fs/promises';
import {createRequire} from 'node:module';
import {spawnSync} from 'node:child_process';
import path from 'node:path';
import {fileURLToPath, pathToFileURL} from 'node:url';
import test from 'node:test';

const require = createRequire(import.meta.url);
const mdxRequire = createRequire(require.resolve('@docusaurus/mdx-loader'));
const imageSizeCjs = mdxRequire('image-size').imageSize;
const {imageSizeFromFile} = mdxRequire('image-size/fromFile');
const imagePackage = path.dirname(path.dirname(mdxRequire.resolve('image-size')));
const metadata = JSON.parse(await readFile(path.join(imagePackage, 'package.json')));
const {imageSize} = await import(pathToFileURL(path.join(imagePackage, metadata.exports['.'].import.default)));
for (const [name, width, height] of [
  ['sidecar-logo.png', 1282, 404],
  ['sidecar-social-card.jpg', 1200, 630],
  ['logo.svg', 200, 200],
]) {
  test(`preserves actual ${name} dimensions through buffer and file APIs`, async () => {
    const filename = fileURLToPath(new URL(`../static/img/${name}`, import.meta.url));
    const bytes = await readFile(filename);
    for (const size of [imageSize(bytes), imageSizeCjs(bytes), await imageSizeFromFile(filename)]) {
      assert.equal(size.width, width);
      assert.equal(size.height, height);
    }
  });
}

// A child timeout bounds malformed parsing even if a dependency regresses.
const parserFixture = `
const kind=process.argv[2];
const parser=require(process.argv[1])[kind.toUpperCase()];
const input=new Uint8Array(kind==='icns'?16:kind==='heif'?48:32);
const view=new DataView(input.buffer);
const box=(offset,size,type)=>{view.setUint32(offset,size,false);input.set(new TextEncoder().encode(type),offset+4)};
if(kind==='icns'){input.set(new TextEncoder().encode('icns'));view.setUint32(4,input.length,false);input.set(new TextEncoder().encode('ic07'),8)}
if(kind==='heif'){box(0,48,'meta');box(12,36,'iprp');box(20,28,'ipco');box(28,0,'ispe');view.setUint32(40,1,false);view.setUint32(44,1,false)}
if(kind==='jxl'){box(0,20,'ftyp');input.set(new TextEncoder().encode('jxl '),8);box(20,0,'jxlp')}
try {console.log(JSON.stringify({size:parser.calculate(input)}))}
catch(error){console.log(JSON.stringify({error:error.message}))}
`;
for (const format of ['icns', 'heif', 'jxl']) {
  test(`terminates malformed ${format} parsing with the expected result`, () => {
    const result = spawnSync(process.execPath,
      ['-e', parserFixture, mdxRequire.resolve(`image-size/types/${format}`), format],
      {encoding: 'utf8', timeout: 2000});
    assert.equal(result.error, undefined, 'parser must terminate without timeout');
    assert.equal(result.status, 0, result.stderr);
    const parsed = JSON.parse(result.stdout);
    if (format === 'icns') assert.equal(parsed.error, 'Invalid ICNS entry length');
    if (format === 'jxl') assert.equal(parsed.error, 'Reached end of input');
    if (format === 'heif') assert.deepEqual(parsed.size, {width: 1, height: 1, type: '\0\0\0\0'});
  });
}

test('retains the CJS APIs used by webpack, SockJS and Express', () => {
  const serialize = createRequire(require.resolve('copy-webpack-plugin'))('serialize-javascript');
  assert.equal(typeof serialize, 'function');
  assert.equal(serialize({text: '<script>'}), '{"text":"\\u003Cscript\\u003E"}');
  assert.match(serialize({pattern: /image/gi}), /new RegExp/);
  const uuid = createRequire(require.resolve('sockjs'))('uuid');
  const identifier = uuid.v4();
  assert.equal(uuid.validate(identifier), true);
  assert.equal(uuid.version(identifier), 4);
  const qs = createRequire(require.resolve('express'))('qs');
  assert.deepEqual(qs.parse('a[b]=one&c=two'), {a: {b: 'one'}, c: 'two'});
  assert.equal(qs.stringify({a: {b: 'one'}}), 'a%5Bb%5D=one');
  assert.deepEqual(qs.parse('a[__proto__][polluted]=yes'), {});
  assert.equal(Object.prototype.polluted, undefined);
});
