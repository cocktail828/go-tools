package z

func ChainCall[T any](fs ...func(T) error) func(T) error {
	return func(in T) error {
		for _, f := range fs {
			if err := f(in); err != nil {
				return err
			}
		}
		return nil
	}
}
