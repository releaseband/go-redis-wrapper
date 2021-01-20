package internal

type CommandPostfixes struct {
	rPush  *keyPostfix
	lTrim  *keyPostfix
	lRange *keyPostfix
}

func NewCommandPostfixes(shardCount int) *CommandPostfixes {
	return &CommandPostfixes{
		rPush:  newKeyPostfix(shardCount),
		lTrim:  newKeyPostfix(shardCount),
		lRange: newKeyPostfix(shardCount),
	}
}

func (c *CommandPostfixes) RPushKey() string {
	return c.rPush.Next()
}

func (c *CommandPostfixes) LRangeKey() string {
	return c.lRange.Next()
}

func (c *CommandPostfixes) LTrimKey() string {
	return c.lTrim.Next()
}
