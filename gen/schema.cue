metaValue: bool | string | number
bodyData: "float" | "double" | "int32" | "uint32" | "int64" | "uint64" | "bool" | "vector2" | "vector3" | "string"
keyName: =~"^[A-Z][A-Za-z0-9_]+$"

#Structure: {
	Meta: [keyName]: metaValue
	Body: [keyName]: bodyData
}

#Structures: {
	list: [keyName]: #Structure
}

structures: #Structures