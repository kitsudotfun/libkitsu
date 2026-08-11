package internal

func IKtShutdown() error {
	if sessionToken == "" {
		return nil
	}

	err := cm.shutdown()
	if err != nil {
		return err
	}

	sessionToken = ""
	attestToken = ""

	return nil
}
