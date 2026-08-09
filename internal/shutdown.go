package internal

func IKtShutdown() error {
	if sessionToken == "" {
		return nil
	}

	err := cm.Shutdown()
	if err != nil {
		return err
	}

	sessionToken = ""
	attestToken = ""

	return nil
}
