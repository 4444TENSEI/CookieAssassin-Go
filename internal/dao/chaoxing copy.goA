package dao

import "encoding/json"

// 查询可以续期的id列表数据
func (br *customRepo[T]) FindChaoxingList(chaoxingID string) ([]string, []string, error) {
	var renewableIDs []string
	var nonRenewableIDs []string
	// 查找处于"续期中"状态的ID
	if err := br.DB.Model(new(T)).
		Select("id").
		Where("chaoxing_id = ? AND update_status = ?", chaoxingID, "续期中").
		Scan(&renewableIDs).Error; err != nil {
		return nil, nil, err
	}
	// 查找"不可续期"状态的ID
	if err := br.DB.Model(new(T)).
		Select("id").
		Where("chaoxing_id = ? AND (update_status = ? OR update_status IS NULL)", chaoxingID, "不可续期").
		Scan(&nonRenewableIDs).Error; err != nil {
		return nil, nil, err
	}
	return renewableIDs, nonRenewableIDs, nil
}

type DataWithID struct {
	ID   uint            `json:"id"`
	Data json.RawMessage `json:"data"`
}

// 根据分类和续期状态，查询可以续期的数据
func (br *customRepo[T]) FindChaoxingTaskIds(queryType string) ([]DataWithID, error) {
	var dataList []DataWithID
	var err error
	baseQuery := br.DB.Model(new(T)).
		Select("id, data").
		Where("type = ?", "超星")
	switch queryType {
	case "all":
		// 查询所有“超星”类型的数据
		err = baseQuery.Scan(&dataList).Error
	case "new":
		// 查询所有“超星”类型、且update_status为空或“不可续期”的数据
		err = baseQuery.Where("update_status IS NULL OR update_status = ?", "不可续期").
			Scan(&dataList).Error
	case "day":
		// 查询所有“超星”类型、且update_status为“续期中”的数据
		err = baseQuery.Where("update_status = ?", "续期中").
			Scan(&dataList).Error
	}
	if err != nil {
		return nil, err
	}
	return dataList, nil
}
