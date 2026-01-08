# Image Repair and Checking

This document describes how to check for missing images and repair them.

## Problem

Sometimes images may be missing from the server even though the database has records of them. This can happen due to:
- Failed downloads during initial sync
- Manual deletion of image files
- Disk space issues
- Network timeouts

## Solution

The service provides two utilities to handle missing images:

1. **pf-check**: Check for missing images (read-only)
2. **pf-repair**: Check and automatically repair missing images

## Commands

### 1. Check for Missing Images

Check which images are missing without making any changes:

```bash
# In Docker container
docker exec -it pf-service /app/pf-check

# Or if running locally
./pf-check
```

**Output example:**
```
PF IMAGE CHECK MODE
Checking for missing images...
✗ Found 5 missing images:

Missing images by property:
  Property ID 123 (pf_id: pf-listing-001): 2 missing images
  Property ID 456 (pf_id: pf-listing-002): 3 missing images

To repair missing images, run: pf_repair
```

### 2. Repair Missing Images

Automatically check for missing images and re-download them:

```bash
# In Docker container
docker exec -it pf-service /app/pf-repair

# Or if running locally
./pf-repair
```

**What it does:**
1. Scans all property images in the database
2. Checks if image files exist on disk
3. For missing images:
   - Fetches the property listing from Property Finder API
   - Gets the original image URLs
   - Re-downloads the missing images
   - Updates the database with new image paths

**Output example:**
```
PF IMAGE REPAIR STARTED...
Checking for missing images...
Found 5 missing images. Starting repair...
Fetching listings for 2 properties...
Fetched 150 listings from API
Repairing 2 missing images for property 123 (pf_id: pf-listing-001)
Successfully repaired image for property 123: property_images/image1.jpg
Successfully repaired image for property 123: property_images/image2.jpg
Repairing 3 missing images for property 456 (pf_id: pf-listing-002)
Successfully repaired image for property 456: property_images/image3.jpg
Successfully repaired image for property 456: property_images/image4.jpg
Successfully repaired image for property 456: property_images/image5.jpg
REPAIR FINISHED: 5 images repaired, 0 failed
```

## How It Works

### Image Checking Process

1. **Database Scan**: Queries all property images from `core_app_propertyimage` table
2. **File System Check**: For each image path, checks if file exists at `MEDIA_ROOT/image_path`
3. **Report**: Lists all missing images grouped by property

### Image Repair Process

1. **Identify Missing Images**: Same as check process
2. **Fetch Property Listings**: Gets all listings from Property Finder API to find image URLs
3. **Match Properties**: Matches database properties with API listings using `pf_id`
4. **Re-download**: For each missing image:
   - Deletes the old database record
   - Downloads the image from Property Finder API
   - Saves new image file
   - Creates new database record

## Configuration

The repair process uses the same configuration as the main sync:

- `PF_API_URL`: Property Finder API URL
- `PF_API_KEY`: API key
- `PF_API_SECRET`: API secret
- `POSTGRES_DSN`: Database connection string
- `MEDIA_ROOT`: Media files root directory (default: `/mhp/media`)
- `IMAGE_DOWNLOAD_MAX_RETRIES`: Max retry attempts (default: 3)
- `IMAGE_DOWNLOAD_RETRY_DELAY`: Delay between retries in seconds (default: 2)
- `IMAGE_DOWNLOAD_TIMEOUT`: Timeout per download in seconds (default: 10)

## Usage Examples

### Check images before repair:
```bash
docker exec -it pf-service /app/pf-check
```

### Repair missing images:
```bash
docker exec -it pf-service /app/pf-repair
```

### Schedule regular checks (add to crontab):
```bash
# Check for missing images daily at 2 AM
0 2 * * * docker exec pf-service /app/pf-check >> /var/log/pf-check.log 2>&1

# Repair missing images weekly on Sunday at 3 AM
0 3 * * 0 docker exec pf-service /app/pf-repair >> /var/log/pf-repair.log 2>&1
```

## Limitations

1. **API Rate Limits**: The repair process fetches all listings from the API, which may be rate-limited
2. **Image Matching**: If the number of images changed in the API, the repair may not perfectly match original images
3. **Missing Listings**: If a property is no longer available in the API, its images cannot be repaired
4. **URL Changes**: If Property Finder changed image URLs, old URLs may no longer work

## Troubleshooting

### "No missing images found" but images are still missing
- Check `MEDIA_ROOT` environment variable matches actual media directory
- Verify database image paths are relative (e.g., `property_images/file.jpg`)
- Check file permissions on media directory

### "Listing not found in API response"
- Property may have been removed from Property Finder
- `pf_id` in database may not match API listing ID
- API may have pagination limits

### "Failed to download image"
- Check network connectivity
- Verify API credentials are correct
- Check `IMAGE_DOWNLOAD_TIMEOUT` is sufficient
- Review retry settings

## Integration with Main Sync

The main sync process (`pf-sync`) continues to work as before. The repair utilities are separate commands that can be run independently:

- **pf-sync**: Regular sync of properties and images
- **pf-check**: Check for missing images (read-only)
- **pf-repair**: Repair missing images (modifies database and downloads files)

## Files

- `cmd/pf_check/main.go`: Check-only command
- `cmd/pf_repair/main.go`: Repair command
- `internal/db/check_images.go`: Database functions for checking images
- `internal/media_download/checker.go`: File system checking utilities

