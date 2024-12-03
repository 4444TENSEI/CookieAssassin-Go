package dao

import "encoding/json"

// 1.查询超星id下所有的cookie id
func (br *customRepo[T]) FindChaoxingList(chaoxingID string) ([]string, error) {
	var allIDs []string
	if err := br.DB.Model(new(T)).
		Select("id").
		Where("chaoxing_id = ?", chaoxingID).
		Scan(&allIDs).Error; err != nil {
		return nil, err
	}
	return allIDs, nil
}

// 2.传入id数组，返回“续期中”的数据切片
func (br *customRepo[T]) FindRenewableDataByID(ids []string) ([]json.RawMessage, error) {
	var dataSlice []json.RawMessage
	if err := br.DB.Model(new(T)).
		Select("data").
		Where("id IN (?) AND update_status = ?", ids, "续期中").
		Scan(&dataSlice).Error; err != nil {
		return nil, err
	}
	return dataSlice, nil
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
