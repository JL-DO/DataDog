package snmp

import (
	"fmt"
)

type valueStoreType struct {
	scalarValues scalarResultValuesType
	columnValues columnResultValuesType
}

// getScalarValues look for oid and returns the value and boolean
// weather valid value has been found
func (v *valueStoreType) getScalarValues(oid string) (snmpValueType, error) {
	value, ok := v.scalarValues[oid]
	if !ok {
		return snmpValueType{}, fmt.Errorf("value for Scalar OID not found: %s", oid)
	}
	return value, nil
}

func (v *valueStoreType) getColumnValues(oid string) (map[string]snmpValueType, error) {
	retValues := make(map[string]snmpValueType)
	values, ok := v.columnValues[oid]
	if !ok {
		return nil, fmt.Errorf("value for Column OID not found: %s", oid)
	}
	for index, value := range values {
		retValues[index] = value
	}

	return retValues, nil
}
