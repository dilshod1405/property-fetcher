# Testlarni ishga tushirish

## 1. Unit testlar (media download)

```bash
cd pfservice
go test ./internal/media_download/... -v
```

## 2. Integration testlar

```bash
cd pfservice
go test -v -run TestSyncDoesNotDeleteDatabaseRecords
go test -v -run TestSyncImageExistenceCheck
```

## 3. Barcha testlarni ishga tushirish

```bash
cd pfservice
go test ./... -v
```

## 4. Docker container ichida testlarni ishga tushirish

```bash
docker exec pf-service sh -c "cd /app && go test ./... -v"
```

## Database o'chirish operatsiyalari tekshiruvi

✅ **pf_sync/main.go** - Hech qanday delete operatsiyasi yo'q
✅ **pf_sync** faqat:
   - `SaveOrUpdateUser` - yangilash yoki yaratish
   - `SaveOrUpdateProperty` - yangilash yoki yaratish  
   - `SavePropertyImage` - yangilash yoki yaratish
   - `CheckMissingImages` - faqat o'qish (read-only)

❌ **pf_sync** hech qachon:
   - `DeletePropertyImage` chaqirmaydi
   - `Delete` yoki `Remove` ishlatmaydi
   - Database recordlarni o'chirmaydi

## Test natijalari

Testlar quyidagilarni tekshiradi:
1. ✅ Har bir image uchun noyob filename yaratiladi
2. ✅ Database recordlar o'chirilmaydi
3. ✅ Missing image check faqat o'qish operatsiyasi (read-only)
4. ✅ Yangi imagelar qo'shiladi, eskilar o'chirilmaydi
