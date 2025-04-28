package semaphore

/*
// Decrement ...
func Decrement(sema string) (int, error) {
	cmd := "scalebox semaphore decrement " + sema
	code, stdout, stderr, err := exec.RunReturnAll(cmd, 20)
	fmt.Printf("stdout:\n%s\n", stdout)
	if err != nil {
		logrus.Errorf("cmd:%s, stderr:\n%s\n", cmd, stderr)
		return math.MinInt, err
	}
	if code > 0 {
		return code, fmt.Errorf("[ERROR]semaphore decrement")
	}
	v, err := strconv.Atoi(strings.TrimSpace(stdout))
	if err != nil {
		logrus.Errorf("semaphore-value not a integer, value=%s\n", stdout)
		return -1, err
	}
	return v, nil
}
*/
