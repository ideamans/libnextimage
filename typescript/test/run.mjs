// テストを走らせ、1件も走らなかったら失敗にする。
//
// 以前は `node --test dist/test/*.test.js` だったが、tsconfig.json は src/ しかコンパイルしないため
// dist/test/ は存在せず、node --test は「対象 0 件・成功」で終わっていた（CI も publish も素通り）。
// テストは tsconfig.test.json で dist-test/ にコンパイルし、ここで件数を確かめる。
import { spawnSync } from 'node:child_process'
import { readFileSync, rmSync } from 'node:fs'

const tap = 'dist-test/report.tap'
rmSync(tap, { force: true })

// グロブは node 自身に展開させる（シェルに任せると Windows で展開されない）
const r = spawnSync(
  process.execPath,
  [
    '--test',
    '--test-reporter=spec',
    '--test-reporter-destination=stdout',
    '--test-reporter=tap',
    `--test-reporter-destination=${tap}`,
    'dist-test/test/**/*.test.js',
  ],
  { stdio: 'inherit' },
)
if (r.status !== 0) process.exit(r.status ?? 1)

const report = readFileSync(tap, 'utf8')
const count = (key) => Number((report.match(new RegExp(`^# ${key} (\\d+)$`, 'm')) || [])[1] ?? NaN)
const tests = count('tests')
const pass = count('pass')
if (!(tests > 0)) {
  console.error(`::error::テストが1件も走らなかった（tests=${tests}）`)
  process.exit(1)
}
if (pass !== tests) {
  console.error(`::error::成功したテストが全体と一致しない（pass=${pass} / tests=${tests}）`)
  process.exit(1)
}
console.log(`${tests} 件のテストが走り、すべて成功した`)
