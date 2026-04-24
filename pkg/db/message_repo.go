package db

import "im-system/proto"

// SaveMessage 持久化一条消息
func SaveMessage(msg *proto.Message) error {
	row := &Message{
		ID:         msg.MsgID,
		MsgType:    msg.MsgType,
		TargetType: msg.TargetType,
		FromUID:    msg.FromUID,
		ToID:       msg.ToID,
		Content:    msg.Content,
		Timestamp:  msg.Timestamp,
	}
	return DB.Create(row).Error
}

// GetHistory 拉取两个用户之间的历史消息（单聊）
func GetHistory(uid1, uid2 int64, lastMsgID int64, limit int) ([]*proto.Message, error) {
	var rows []Message
	query := DB.Where(
		"target_type = ? AND ((from_uid = ? AND to_id = ?) OR (from_uid = ? AND to_id = ?))",
		proto.TargetTypeSingle, uid1, uid2, uid2, uid1,
	).Order("id asc").Limit(limit)

	if lastMsgID > 0 {
		query = query.Where("id > ?", lastMsgID)
	}
	if err := query.Find(&rows).Error; err != nil {
		return nil, err
	}

	msgs := make([]*proto.Message, len(rows))
	for i, r := range rows {
		msgs[i] = &proto.Message{
			MsgID:      r.ID,
			MsgType:    r.MsgType,
			TargetType: r.TargetType,
			FromUID:    r.FromUID,
			ToID:       r.ToID,
			Content:    r.Content,
			Timestamp:  r.Timestamp,
		}
	}
	return msgs, nil
}

// GetGroupHistory 拉取群聊历史消息
func GetGroupHistory(groupID int64, lastMsgID int64, limit int) ([]*proto.Message, error) {
	var rows []Message
	query := DB.Where("target_type = ? AND to_id = ?", proto.TargetTypeGroup, groupID).
		Order("id asc").Limit(limit)

	if lastMsgID > 0 {
		query = query.Where("id > ?", lastMsgID)
	}
	if err := query.Find(&rows).Error; err != nil {
		return nil, err
	}

	msgs := make([]*proto.Message, len(rows))
	for i, r := range rows {
		msgs[i] = &proto.Message{
			MsgID:      r.ID,
			MsgType:    r.MsgType,
			TargetType: r.TargetType,
			FromUID:    r.FromUID,
			ToID:       r.ToID,
			Content:    r.Content,
			Timestamp:  r.Timestamp,
		}
	}
	return msgs, nil
}
