package NoQ_RoomQ

type validationResult struct {
	redirectURL string
}

func ValidationResult(redirectURL string) validationResult {
	return validationResult{redirectURL: redirectURL}
}

func (vr validationResult) NeedRedirect() bool {
	return len(vr.redirectURL) > 0
}

func (vr validationResult) GetRedirectURL() string {
	return vr.redirectURL
}
