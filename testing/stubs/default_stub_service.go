package stubs

type DefaultStubService struct {
	ReturnObj    interface{}
	Err          error
	MethodCalled []map[string]interface{}
}

func (s *DefaultStubService) CalledCnt() int {
	return len(s.MethodCalled)
}

func (s *DefaultStubService) HaveBeenCalledWith(matchingFnc func(args map[string]any) bool) bool {
	r := false

	for _, args := range s.MethodCalled {
		if matchingFnc(args) {
			r = true
		}
	}

	return r
}
