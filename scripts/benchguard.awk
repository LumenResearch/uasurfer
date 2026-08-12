# Reads a `benchstat -format csv` comparison of master against a change and fails
# on a regression that matters.
#
# Two conditions, both required, because either on its own gets it wrong:
#
#   thresh  the slowdown as a percentage. On its own it fails a check that went
#           from 15ns to 85ns, which is a rounding error inside a parse that
#           still costs half a microsecond.
#   minns   how many nanoseconds slower, absolutely. On its own it says nothing
#           about whether that was a lot, and a slow CI runner inflates every
#           number it sees anyway - which is also why this compares two runs on
#           the same machine rather than measuring against a fixed ceiling.
#
# So a parse has to be both meaningfully and measurably slower to fail. What that
# catches is the change of shape - a regexp on the hot path, a scan gone
# quadratic - and not the nanoseconds anyone should be free to spend on accuracy.
#
# Time only. The comparison is run without -benchmem, so allocations are not in
# this input; `make bench` is where you see those.
#
#   benchstat -format csv base.out head.out | awk -v thresh=30 -v minns=300 -f benchguard.awk

BEGIN {
	FS = ","
	if (thresh == "") thresh = 30
	if (minns == "") minns = 300
	bad = 0
	rows = 0
}

/^,.*sec\/op,CI/ { metric = "time"; next }

# name, old, CI, new, CI, delta, p
/^[A-Za-z]/ {
	name = $1
	if (metric != "time") next
	old = $2 + 0
	new = $4 + 0
	delta = $6
	rows++

	if (delta !~ /^\+/) next  # "~" for no significant difference, or a decrease

	slower = (new - old) * 1e9  # benchstat reports seconds
	if (delta + 0 > thresh && slower > minns) {
		printf "%s: %s slower, %+.0f ns/op, past %d%% and %d ns\n", name, delta, slower, thresh, minns
		bad = 1
	} else if (delta + 0 > thresh) {
		printf "note: %s is %s slower, but only %+.0f ns/op\n", name, delta, slower
	}
}

END {
	if (rows == 0) {
		print "no benchmarks compared"
		exit 1
	}
	if (bad) exit 1
	printf "%d benchmarks compared, none more than %d%% and %d ns slower\n", rows, thresh, minns
}
