package dao

// 单条更新-指定3个参数：要更新的目标id、目标id数据下的字段名、新的值
func (br *customRepo[T]) Update(id uint, field string, value interface{}) error {
	return br.DB.Model(new(T)).Where("id = ?", id).Update(field, value).Error
}

// 条件更新
// 1.指定字段名和值的字典(值如果不唯一会导致更新多条符合的数据) 2.要更新的字段和对应的值
// 使用示例请全局搜索函数: UpdateProfile
func (br *customRepo[T]) UpdateByField(field string, value interface{}, updateList map[string]interface{}) error {
	return br.DB.Model(new(T)).Where(field+" = ?", value).Updates(updateList).Error
}
