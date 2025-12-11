package property

import "gorm.io/gorm"

func SaveEnglishTranslation(db *gorm.DB, propID uint, title, addr, desc string) {
	db.Exec(`
        INSERT INTO core_app_property_translation (master_id, language_code, title, address, description)
        VALUES (?, 'en', ?, ?, ?)
        ON CONFLICT (master_id, language_code) DO UPDATE 
        SET title = EXCLUDED.title, 
            description = EXCLUDED.description,
            address = EXCLUDED.address
    `, propID, title, addr, desc)
}
