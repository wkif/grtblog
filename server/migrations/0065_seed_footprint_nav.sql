-- +goose Up

INSERT INTO nav_menu (name, url, icon, sort, parent_id, created_at, updated_at)
SELECT '足迹', '/footprints', 'map-pin', 31, parent.id, NOW(), NOW()
FROM nav_menu parent
WHERE parent.name = '更多'
  AND parent.parent_id IS NULL
  AND parent.deleted_at IS NULL
  AND NOT EXISTS (
      SELECT 1 FROM nav_menu existing
      WHERE existing.url = '/footprints' AND existing.deleted_at IS NULL
  );

-- +goose Down

DELETE FROM nav_menu
WHERE url = '/footprints';
