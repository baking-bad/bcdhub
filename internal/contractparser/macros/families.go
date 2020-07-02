package macros

// GetAllFamilies -
func GetAllFamilies() *[]Family {
	return &[]Family{
		failFamily{},
		ifLeftFamily{},
		ifNoneFamily{},
		unpairFamily{},
		cadrFamily{},
		setCarFamily{},
		setCdrFamily{},
		mapFamily{},
		ifFamily{},
	}
}

// GetLangugageFamilies -
func GetLangugageFamilies() *[]Family {
	return &[]Family{
		failFamily{},
		ifLeftFamily{},
		ifNoneFamily{},
		unpairFamily{},
		setCarFamily{},
		setCdrFamily{},
		mapFamily{},
		ifFamily{},
	}
}
