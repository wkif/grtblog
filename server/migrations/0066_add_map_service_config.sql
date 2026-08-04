-- +goose Up
INSERT INTO sys_config (
    config_key,
    value,
    is_sensitive,
    group_path,
    label,
    description,
    value_type,
    enum_options,
    visible_when,
    sort,
    meta
)
VALUES
    (
        'map.provider',
        'osm',
        false,
        'map/service',
        '地图服务商',
        'OpenStreetMap 无需 Key；天地图需要配置开发者 Key。',
        'enum',
        '[{"label":"OpenStreetMap","value":"osm"},{"label":"天地图","value":"tianditu"}]'::jsonb,
        '[]'::jsonb,
        10,
        '{}'::jsonb
    ),
    (
        'map.tianditu.key',
        '',
        false,
        'map/tianditu',
        '天地图 Key',
        '从天地图开发者控制台申请。Key 会随浏览器中的地图瓦片请求发送，请同时配置域名白名单。',
        'string',
        '[]'::jsonb,
        '[{"key":"map.provider","op":"eq","value":"tianditu"}]'::jsonb,
        20,
        '{"inputType":"password"}'::jsonb
    ),
    (
        'map.tianditu.layer',
        'vector',
        false,
        'map/tianditu',
        '底图类型',
        '选择天地图矢量底图或卫星影像底图，均叠加中文注记。',
        'enum',
        '[{"label":"矢量地图","value":"vector"},{"label":"卫星影像","value":"imagery"}]'::jsonb,
        '[{"key":"map.provider","op":"eq","value":"tianditu"}]'::jsonb,
        30,
        '{}'::jsonb
    )
ON CONFLICT (config_key) DO NOTHING;

-- +goose Down
DELETE FROM sys_config
WHERE config_key IN ('map.provider', 'map.tianditu.key', 'map.tianditu.layer');
