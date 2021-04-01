package version

func init() {
	// The main version number that is being run at the moment.
	// 版本迭代 从1.1.5原版号开始
	Version = "1.1.7"

	// A pre-release marker for the version. If this is "" (empty string)
	// then it means that it is a final release. Otherwise, this is a pre-release
	// such as "dev" (in development), "beta", "rc1", "stable", etc.
	VPrerelease = "stable"
}
