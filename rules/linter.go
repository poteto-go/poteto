//go:build ruleguard
// +build ruleguard

package rules

import "github.com/quasilyte/go-ruleguard/dsl"

func boolFunctionNaming(m dsl.Matcher) {
	m.Match(`func $name($*params) bool { $*body }`).
		Where(!m["name"].Text.Matches(`^(Is|is|Has|has).*`)).
		Report("bool function name should start with 'Is' | 'is' | 'Has' | 'has'")
}

func boolExprSimplify(m dsl.Matcher) {
	m.Match(`!($x != $y)`).Report(`$x == $y`)
	m.Match(`!($x == $y)`).Report(`$x != $y`)
}

func exposedMutex(m dsl.Matcher) {
	isExported := func(v dsl.Var) bool {
		return v.Text.Matches(`^\p{Lu}`)
	}

	m.Match(`type $name struct { $*_; sync.Mutex; $*_ }`).
		Where(isExported(m["name"])).
		Report("do not embed sync.Mutex")

	m.Match(`type $name struct { $*_; sync.RWMutex; $*_ }`).
		Where(isExported(m["name"])).
		Report("do not embed sync.RWMutex")
}

func timeEqual(m dsl.Matcher) {
	m.Match("$t0 == $t1").Where(m["t0"].Type.Is("time.Time")).Report("using == with time.Time")
	m.Match("$t0 != $t1").Where(m["t0"].Type.Is("time.Time")).Report("using != with time.Time")
	m.Match(`map[$k]$v`).Where(m["k"].Type.Is("time.Time")).Report("map with time.Time keys are easy to misuse")
}

func oddCompoundOp(m dsl.Matcher) {
	m.Match("$x += $x + $_",
		"$x += $x - $_").
		Report("odd += expression")

	m.Match("$x -= $x + $_",
		"$x -= $x - $_").
		Report("odd -= expression")

	m.Match("$x *= $x * $_",
		"$x *= $x / $_").
		Report("odd *= expression")

	m.Match("$x /= $x * $_",
		"$x /= $x / $_").
		Report("odd /= expression")
}

func oddComparisons(m dsl.Matcher) {
	m.Match("$x - $y == 0").
		Report("odd comparison").
		Suggest("$x == $y")

	m.Match("$x - $y != 0").
		Report("odd comparison").
		Suggest("$x != $y")

	m.Match("$x - $y < 0").
		Report("odd comparison").
		Suggest("$y > $x")

	m.Match("$x - $y <= 0").
		Report("odd comparison").
		Suggest("$y >= $x")

	m.Match("$x - $y > 0").
		Report("odd comparison").
		Suggest("$x > $y")

	m.Match("$x - $y >= 0").
		Report("odd comparison").
		Suggest("$x >= $y")

	m.Match("$x ^ $y == 0").
		Report("odd comparison").
		Suggest("$x == $y")

	m.Match("$x ^ $y != 0").
		Report("odd comparison").
		Suggest("$x != $y")
}

func sprintErr(m dsl.Matcher) {
	m.Match(`fmt.Sprint($err)`,
		`fmt.Sprintf("%s", $err)`,
		`fmt.Sprintf("%v", $err)`,
	).
		Where(m["err"].Type.Is("error")).
		Report("maybe call $err.Error() instead of fmt.Sprint()?")
}

func nilErr(m dsl.Matcher) {
	m.Match(
		`if err == nil { return err }`,
		`if err == nil { return $*_, err }`,
	).
		Report(`return nil error instead of nil value`)
}

func lenStrByteSlice(m dsl.Matcher) {
	// len(string([]byte)) -> len([]byte)
	m.Match(`len(string($b))`).
		Where(m["b"].Type.Underlying().Is("[]byte")).
		Report(`Call len() on the byte slice instead of converting to a string first`).
		Suggest(`len($b)`)
}

func lenByteSliceStr(m dsl.Matcher) {
	// len([]byte(string)) -> len(string)
	m.Match(`len([]byte($s))`).
		Where(m["s"].Type.Underlying().Is("string")).
		Report(`Call len() on the string instead of converting to []byte first.`).
		Suggest(`len($s)`)
}

func badLock(m dsl.Matcher) {
	// `mu1` and `mu2` are added to make possible report a line where `m2` is used (with a defer)

	// no defer
	m.Match(`$mu1.Lock(); $mu2.Unlock()`).
		Where(m["mu1"].Text == m["mu2"].Text).
		Report(`defer is missing, mutex is unlocked immediately`).
		At(m["mu2"])

	m.Match(`$mu1.RLock(); $mu2.RUnlock()`).
		Where(m["mu1"].Text == m["mu2"].Text).
		Report(`defer is missing, mutex is unlocked immediately`).
		At(m["mu2"])

	// different lock operations
	m.Match(`$mu1.Lock(); defer $mu2.RUnlock()`).
		Where(m["mu1"].Text == m["mu2"].Text).
		Report(`suspicious unlock, maybe Unlock was intended?`).
		At(m["mu2"])

	m.Match(`$mu1.RLock(); defer $mu2.Unlock()`).
		Where(m["mu1"].Text == m["mu2"].Text).
		Report(`suspicious unlock, maybe RUnlock was intended?`).
		At(m["mu2"])

	// double locks
	m.Match(`$mu1.Lock(); defer $mu2.Lock()`).
		Where(m["mu1"].Text == m["mu2"].Text).
		Report(`maybe defer $mu1.Unlock() was intended?`).
		At(m["mu2"])

	m.Match(`$mu1.RLock(); defer $mu2.RLock()`).
		Where(m["mu1"].Text == m["mu2"].Text).
		Report(`maybe defer $mu1.RUnlock() was intended?`).
		At(m["mu2"])
}
