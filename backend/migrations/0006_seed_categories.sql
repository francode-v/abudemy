-- 0006_seed_categories.sql
-- Baseline service categories so the catalogue is usable out of the box.

INSERT INTO service_categories (name, slug, description) VALUES
    ('Salud y bienestar',        'salud-bienestar',   'Centros de salud, consultas y programas de bienestar.'),
    ('Asistencia social',        'asistencia-social', 'Apoyo social, acompañamiento y ayuda domiciliaria.'),
    ('Actividades comunitarias', 'actividades',       'Clubes, talleres y encuentros comunitarios.'),
    ('Educación y cursos',       'educacion',         'Cursos, capacitaciones y alfabetización digital.'),
    ('Trámites y apoyo legal',   'tramites',          'Orientación para trámites y asesoría legal.'),
    ('Comercio y servicios',     'comercio',          'Comercios y servicios de proximidad.'),
    ('Transporte y movilidad',   'transporte',        'Opciones de transporte y acompañamiento.')
ON CONFLICT (slug) DO NOTHING;
