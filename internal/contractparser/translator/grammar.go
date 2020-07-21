package translator

const defaultGrammar = `
	instrs         <- instr (Semi instr)* Semi?    
	instr          <- expr / Byte / Int / String / complex_instr
	complex_instr  <- (LeftCurly instrs* RightCurly)
	expr           <- prim annots? args?
	annots         <- (annot+)
	annot          <- < Annot >
	args           <- (arg+)
	arg            <- prim / Byte / Int / String / complex_instr / (LeftParen expr RightParen)
	prim           <- < Alpha Accessable+ >
	
	Int        <- < Minus? Digit+ >
	Byte       <- < HexPrefix Hex+ >
    String     <- < '"' StringContent* '"' >
    Annot      <- AnnotPrefix+ AnnotBody*
    LeftParen  <- '('
	RightParen <- ')'
	LeftCurly  <- '{'
	RightCurly <- '}'
	Semi       <- ';'
	Minus      <- '-'
	Dot        <- '.'

	
	StringContent <- [_a-zA-Z0-9- /:,\\.'\\\\(\\)\\*\\+><=!\[\]^?%$;#№@~{}]
	Hex           <- [A-F0-9a-f]
	Alpha         <- [a-zA-Z]
	Accessable    <- [A-Za-z0-9_]
	Digit         <- [0-9]
	HexPrefix     <- '0' 'x'
	AnnotPrefix   <- [:@%]
	AnnotBody     <- [_0-9a-zA-Z\\.]

	%whitespace <- [ \t\r\n]*
`
