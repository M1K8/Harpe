package db

import (
	"context"
	"errors"
	"fmt"
	"log"
)

func (d *DB) InitialiseServer(guildID, permID string) error {
	contxt := context.Background()

	res, err := d.GetAllAlerters(guildID)

	if err != nil {
		a := &Channel{
			UserID:        "0",
			RoleID:        "0",
			ChannelID:     "0",
			GuildID:       guildID,
			PermissionsID: permID,
		}

		_, err := d.db.NewInsert().Model(a).On("CONFLICT (user_id) DO UPDATE").Exec(contxt)

		if err != nil {
			log.Println(fmt.Sprintf("Unable to create alerter %v : %v", a, err.Error()))
			return err
		}
		return nil
	}

	for _, v := range res {
		if v.PermissionsID != permID {
			// recreate all alerters with the correct role; assume theyre all dirty
			for _, v2 := range res {
				err = d.CreateAlerter(guildID, v2.ChannelID, v2.UserID, v2.RoleID, permID)
				if err != nil {
					log.Println(err)
				}

			}
			return nil
		}
	}

	return nil
}

func (d *DB) CreateAlerter(guild, channelID, userID, roleID, permID string) error {
	contxt := context.Background()

	if guild != d.Guild {
		return errors.New("Incorrect Guild!")
	}
	a := &Channel{
		UserID:        userID,
		RoleID:        roleID,
		ChannelID:     channelID,
		GuildID:       guild,
		PermissionsID: permID,
	}

	_, err := d.db.NewInsert().Model(a).On("CONFLICT (user_id) DO UPDATE").Exec(contxt)

	if err != nil {
		log.Println(fmt.Sprintf("Unable to create alerter %v : %v", a, err.Error()))
		return err
	}
	return nil

}

func (d *DB) RemoveAlerter(guild, userID string) error {
	contxt := context.Background()

	if guild != d.Guild {
		return errors.New("Incorrect Guild!")
	}
	a := &Channel{
		UserID:  userID,
		GuildID: guild,
	}

	res, err := d.db.NewDelete().Model(a).Where("user_id = ?", userID).Exec(contxt)
	rowsAffected, _ := res.RowsAffected()

	if rowsAffected == 0 {
		err = errors.New(fmt.Sprintf("Unable to remove Alerter %v : NOT FOUND", userID))
		return err
	}
	if err != nil {
		log.Println(fmt.Sprintf("Unable to delete alerter %v : %v", userID, err.Error()))
		return err
	}
	return nil

}

func (d *DB) GetAlerter(guild, userID string) (*Channel, error) {
	if guild != d.Guild {
		return nil, errors.New("Incorrect Guild!")
	}
	contxt := context.Background()
	a := &Channel{
		UserID:  userID,
		GuildID: guild,
	}
	err := d.db.NewSelect().Model(a).Where("user_id = ?", userID).Scan(contxt)

	if err != nil {
		log.Println(fmt.Sprintf("Unable to get alerter %v : %v", userID, err.Error()))
		return nil, err
	}
	return a, nil
}

func (d *DB) GetAllAlerters(guild string) ([]*Channel, error) {
	if guild != d.Guild {
		return nil, errors.New("Incorrect Guild!")
	}
	allAlerters := make([]*Channel, 0)
	contxt := context.Background()
	err := d.db.NewSelect().Model((*Channel)(nil)).Where("guild_id = ?", guild).Scan(contxt, &allAlerters)

	if err != nil {
		log.Println(fmt.Sprintf("Unable to get alerters %v : %v", guild, err.Error()))
		return nil, err
	}
	return allAlerters, nil
}
