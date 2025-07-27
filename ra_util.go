package ra

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func getBaseFlag(flag any) *BaseFlag {
	switch f := flag.(type) {
	case *BoolFlag:
		return &f.BaseFlag
	case *StringFlag:
		return &f.BaseFlag
	case *IntFlag:
		return &f.BaseFlag
	case *Int64Flag:
		return &f.BaseFlag
	case *Float64Flag:
		return &f.BaseFlag
	case *StringSliceFlag:
		return &f.BaseFlag
	case *IntSliceFlag:
		return &f.BaseFlag
	case *Int64SliceFlag:
		return &f.BaseFlag
	case *Float64SliceFlag:
		return &f.BaseFlag
	case *BoolSliceFlag:
		return &f.BaseFlag
	}
	return nil
}
