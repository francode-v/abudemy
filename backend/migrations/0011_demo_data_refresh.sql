-- 0011_demo_data_refresh.sql
-- Pone el dataset de demostración (0007 + 0008) en sintonía con las
-- funcionalidades actuales de la plataforma:
--
--   * Perfil sociodemográfico (0010) de las personas mayores del demo, para
--     que "Personas registradas", sus filtros e indicadores tengan datos.
--   * Flujo de aprobación institucional (0009): un administrador demo y una
--     cuenta institucional pendiente de aprobación.
--   * Actividades, recordatorios y oportunidades re-anclados al futuro y en
--     horario de Chile, para que el demo siga mostrando eventos "próximos"
--     aunque la base se haya creado hace semanas; los títulos estacionales
--     pasan a ser neutros.
--
-- Es idempotente y solo actúa si el dataset demo ya existe: si 0007/0008 no
-- se aplicaron (por ejemplo, en una base de producción sin datos demo), la
-- migración no inserta nada.
--
-- Cuentas demo tras esta migración:
--   admin@abudemy.cl                    / Admin1234      (admin)
--   institucional@abudemy.cl            / Demo1234       (institutional_client aprobado)
--   web.institucional@abudemy.cl        / WebDemo1234    (institutional_client aprobado)
--   pendiente.institucional@abudemy.cl  / Demo1234       (institutional_client pendiente)
--   maria.gonzalez@ejemplo.cl, juan.soto@ejemplo.cl, rosa.munoz@ejemplo.cl,
--   carmen.rojas@ejemplo.cl, elena.torres@ejemplo.cl, patricia.nunez@ejemplo.cl,
--   hugo.fuentes@ejemplo.cl, teresa.aguilera@ejemplo.cl, raul.paredes@ejemplo.cl,
--   silvia.castro@ejemplo.cl            / Demo1234       (mobile_user)
--
-- Las contraseñas corresponden a los hash bcrypt del propio dataset demo
-- (coste 12). Nunca reutilices estas cuentas fuera de demostración.

DO $demo$
BEGIN
    -- Solo refresca bases que ya tienen el dataset de demostración.
    IF NOT EXISTS (
        SELECT 1 FROM users WHERE id = '46c2c334-2936-4b1d-a8e4-9d9388462fee'
    ) THEN
        RAISE NOTICE '0011_demo_data_refresh: dataset demo ausente; no se insertan datos';
        RETURN;
    END IF;

    -- ------------------------------------------------------------------
    -- 1. Perfil sociodemográfico de las personas demo existentes (0010).
    -- Los valores siguen el formulario de registro de la app móvil y dan
    -- variedad a los filtros de "Personas registradas".
    -- ------------------------------------------------------------------
    UPDATE users SET comuna = 'Concepción', health_system = 'fonasa',
        sex = 'female', living_situation = 'with_family', updated_at = now()
    WHERE id = '7a899387-0cb1-4ffa-8875-8029de2030e1'; -- María González

    UPDATE users SET comuna = 'Talcahuano', health_system = 'fonasa',
        sex = 'male', living_situation = 'with_partner', updated_at = now()
    WHERE id = 'b51be962-5721-430d-acb4-5e8127918b2a'; -- Juan Soto

    UPDATE users SET comuna = 'Talcahuano', health_system = 'capredena',
        sex = 'female', living_situation = 'alone', updated_at = now()
    WHERE id = '46c2c334-2936-4b1d-a8e4-9d9388462fee'; -- Rosa Muñoz

    UPDATE users SET comuna = 'Talcahuano', health_system = 'fonasa',
        sex = 'female', living_situation = 'with_children', updated_at = now()
    WHERE id = '0b446edf-17aa-4784-a848-e242d285649a'; -- Carmen Rojas

    UPDATE users SET comuna = 'Concepción', health_system = 'isapre',
        sex = 'female', living_situation = 'with_partner_and_children', updated_at = now()
    WHERE id = '627e1362-ab18-43d4-9cfe-e6bcae22320e'; -- Elena Torres

    -- ------------------------------------------------------------------
    -- 2. Personas adicionales para que los filtros, los indicadores y la
    -- paginación de "Personas registradas" tengan una muestra realista
    -- (comuna, edad, sexo, previsión y convivencia variadas).
    -- ------------------------------------------------------------------
    INSERT INTO users (id, email, password_hash, full_name, phone, birth_date,
        comuna, health_system, sex, living_situation, role, status,
        institutional_approved, last_login_at, created_at, updated_at)
    VALUES
    ('3f0c7b1e-8a52-4c1d-9e6f-1b2c3d4e5f60', 'patricia.nunez@ejemplo.cl',
     '$2a$12$wLYWdCUBGqDmIJIUNEPQjeVz4a.1LoGf4VzoMBU1OfcT.vUdG58qK',
     'Patricia Núñez Salazar', '+56945667788', '1955-06-09',
     'Concepción', 'fonasa', 'female', 'alone', 'mobile_user', 'active',
     true, now() - interval '1 day', now() - interval '38 days', now() - interval '1 day'),

    ('a1b2c3d4-5e6f-4a7b-8c9d-0e1f2a3b4c5d', 'hugo.fuentes@ejemplo.cl',
     '$2a$12$wLYWdCUBGqDmIJIUNEPQjeVz4a.1LoGf4VzoMBU1OfcT.vUdG58qK',
     'Hugo Fuentes Rivas', '+56946778899', '1949-01-30',
     'Hualpén', 'isapre', 'male', 'with_partner', 'mobile_user', 'active',
     true, now() - interval '3 days', now() - interval '35 days', now() - interval '3 days'),

    ('b2c3d4e5-6f7a-4b8c-9d0e-1f2a3b4c5d6e', 'teresa.aguilera@ejemplo.cl',
     '$2a$12$wLYWdCUBGqDmIJIUNEPQjeVz4a.1LoGf4VzoMBU1OfcT.vUdG58qK',
     'Teresa Aguilera Pinto', '+56947889900', '1960-09-14',
     'Chiguayante', 'capredena', 'female', 'with_children', 'mobile_user', 'active',
     true, now() - interval '5 days', now() - interval '30 days', now() - interval '5 days'),

    ('c3d4e5f6-7a8b-4c9d-8e0f-2a3b4c5d6e7f', 'raul.paredes@ejemplo.cl',
     '$2a$12$wLYWdCUBGqDmIJIUNEPQjeVz4a.1LoGf4VzoMBU1OfcT.vUdG58qK',
     'Raúl Paredes Moreno', '+56948990011', '1958-04-22',
     'San Pedro de la Paz', 'fonasa', 'male', 'with_family', 'mobile_user', 'active',
     true, now() - interval '2 days', now() - interval '27 days', now() - interval '2 days'),

    ('d4e5f6a7-8b9c-4d0e-9f1a-3b4c5d6e7f80', 'silvia.castro@ejemplo.cl',
     '$2a$12$wLYWdCUBGqDmIJIUNEPQjeVz4a.1LoGf4VzoMBU1OfcT.vUdG58qK',
     'Silvia Castro Vidal', '+56949001122', '1953-11-05',
     'Talcahuano', 'dipreca', 'female', 'with_caregiver', 'mobile_user', 'active',
     true, now() - interval '6 days', now() - interval '24 days', now() - interval '6 days')
    ON CONFLICT DO NOTHING;

    -- Inscripciones de las personas nuevas en actividades publicadas, para
    -- que las fichas de actividad y los informes tengan más movimiento.
    INSERT INTO activity_enrollments (id, activity_id, user_id, status, created_at, updated_at)
    VALUES
    ('f0e1d2c3-b4a5-4968-8778-99aabbccdde1',
     '0147a25e-41f8-4b9f-a990-96244645b22f', '3f0c7b1e-8a52-4c1d-9e6f-1b2c3d4e5f60',
     'enrolled', now() - interval '2 days', now() - interval '2 days'),
    ('f0e1d2c3-b4a5-4968-8778-99aabbccdde2',
     'b38c372b-07af-448d-a5b7-840869030e6b', 'a1b2c3d4-5e6f-4a7b-8c9d-0e1f2a3b4c5d',
     'enrolled', now() - interval '4 days', now() - interval '4 days'),
    ('f0e1d2c3-b4a5-4968-8778-99aabbccdde3',
     'e87ad0f8-7ea0-4bfb-92ab-17b430419f37', 'b2c3d4e5-6f7a-4b8c-9d0e-1f2a3b4c5d6e',
     'enrolled', now() - interval '5 days', now() - interval '5 days'),
    ('f0e1d2c3-b4a5-4968-8778-99aabbccdde4',
     '866b8521-98cf-48cd-8e9b-eef1154a1095', 'c3d4e5f6-7a8b-4c9d-8e0f-2a3b4c5d6e7f',
     'enrolled', now() - interval '1 day', now() - interval '1 day'),
    ('f0e1d2c3-b4a5-4968-8778-99aabbccdde5',
     'ee35fdf7-4489-4a50-a1d4-0c3ca81a5dbf', 'd4e5f6a7-8b9c-4d0e-9f1a-3b4c5d6e7f80',
     'waitlist', now() - interval '6 days', now() - interval '6 days')
    ON CONFLICT DO NOTHING;

    -- ------------------------------------------------------------------
    -- 3. Flujo de aprobación institucional (0009).
    -- Las cuentas que administran las organizaciones del demo se mantienen
    -- aprobadas; se agrega una cuenta pendiente para poder probar el flujo
    -- de "Cuentas institucionales" y un administrador para el portal.
    -- ------------------------------------------------------------------
    UPDATE users SET institutional_approved = true, status = 'active', updated_at = now()
    WHERE id IN ('8bf68f6e-3b3f-4763-8784-b88ab564db14',
                 '69297821-b7e3-4cad-a4b4-90df97ca1a07');

    INSERT INTO users (id, email, password_hash, full_name, phone, birth_date,
        comuna, health_system, sex, living_situation, role, status,
        institutional_approved, last_login_at, created_at, updated_at)
    VALUES
    ('e5f6a7b8-9c0d-4e1f-8a2b-4c5d6e7f8091', 'pendiente.institucional@abudemy.cl',
     '$2a$12$wLYWdCUBGqDmIJIUNEPQjeVz4a.1LoGf4VzoMBU1OfcT.vUdG58qK',
     'Clínica Los Cerros (solicitud pendiente)', '+56412229988', NULL,
     'Talcahuano', '', '', '', 'institutional_client', 'active',
     false, NULL, now() - interval '12 days', now() - interval '12 days'),

    ('f6a7b8c9-0d1e-4f2a-9b3c-5d6e7f809102', 'admin@abudemy.cl',
     '$2a$12$3HvW7HxvVSR9YtVeT8eiR.OMHWfw7kb0355u2oix/sGNwpzwg9Pd6',
     'Administrador de la plataforma', '', NULL,
     '', '', '', '', 'admin', 'active',
     true, now(), now() - interval '60 days', now())
    ON CONFLICT DO NOTHING;

    -- ------------------------------------------------------------------
    -- 4. Re-anclar el calendario del demo.
    -- Las actividades quedan repartidas en los próximos 24 días, en
    -- horario de Chile (America/Santiago) y con horas variadas, y las
    -- oportunidades en los próximos 3 meses. Así el demo sigue mostrando
    -- eventos "próximos" aunque la base se haya migrado hace semanas.
    -- ------------------------------------------------------------------
    UPDATE activities a SET
        start_at = c.start_at,
        end_at = c.start_at + interval '2 hours'
    FROM (
        SELECT d.id,
               ((date_trunc('day', now() AT TIME ZONE 'America/Santiago')
                 + interval '1 day'
                 + (d.rn * interval '2 days')
                 + ((ARRAY[10, 12, 16, 18])[1 + ((d.rn - 1) % 4)::int] * interval '1 hour')
               ) AT TIME ZONE 'America/Santiago') AS start_at
        FROM (
            SELECT id, row_number() OVER (ORDER BY created_at, id) AS rn
            FROM activities
            WHERE id IN (
                '7268a457-2237-4491-9c5f-5f6805081b02',
                '3ab1d734-b523-464f-809c-503c0aebbcab',
                '0161fad0-25e4-45db-9274-66829c2e636b',
                '727693ca-282a-4cd8-97a0-6cc1ac4b0937',
                'ee35fdf7-4489-4a50-a1d4-0c3ca81a5dbf',
                '4ff693cd-a350-49af-a17c-56ae79478c2c',
                '0147a25e-41f8-4b9f-a990-96244645b22f',
                '7f89ec89-fd09-4d40-99ff-42e7e1ceec6f',
                '866b8521-98cf-48cd-8e9b-eef1154a1095',
                'b38c372b-07af-448d-a5b7-840869030e6b',
                'd7dc6460-47b6-406b-81c7-8f40d5b8bb45',
                'e87ad0f8-7ea0-4bfb-92ab-17b430419f37'
            )
        ) d
    ) c
    WHERE a.id = c.id;

    -- Títulos estacionales que quedaban desfasados con el calendario.
    UPDATE activities SET title = 'Coro comunitario del barrio', updated_at = now()
    WHERE id = '0161fad0-25e4-45db-9274-66829c2e636b';

    UPDATE activities SET title = 'Minga vecinal del barrio', updated_at = now()
    WHERE id = '0147a25e-41f8-4b9f-a990-96244645b22f';

    -- Los recordatorios acompañan al nuevo inicio de su actividad.
    UPDATE activity_reminders r
    SET remind_at = a.start_at - interval '1 day', updated_at = now()
    FROM activities a
    WHERE r.activity_id = a.id
      AND a.id IN (
          '7268a457-2237-4491-9c5f-5f6805081b02',
          '3ab1d734-b523-464f-809c-503c0aebbcab',
          '0161fad0-25e4-45db-9274-66829c2e636b',
          '727693ca-282a-4cd8-97a0-6cc1ac4b0937',
          'ee35fdf7-4489-4a50-a1d4-0c3ca81a5dbf',
          '4ff693cd-a350-49af-a17c-56ae79478c2c',
          '0147a25e-41f8-4b9f-a990-96244645b22f',
          '7f89ec89-fd09-4d40-99ff-42e7e1ceec6f',
          '866b8521-98cf-48cd-8e9b-eef1154a1095',
          'b38c372b-07af-448d-a5b7-840869030e6b',
          'd7dc6460-47b6-406b-81c7-8f40d5b8bb45',
          'e87ad0f8-7ea0-4bfb-92ab-17b430419f37'
      );

    UPDATE opportunities o SET
        expires_at = now() + (d.rn * interval '10 days') + interval '7 days'
    FROM (
        SELECT id, row_number() OVER (ORDER BY created_at, id) AS rn
        FROM opportunities
        WHERE id IN (
            'ca3b2668-06f9-4b8a-a892-23d73f67f453',
            '2ed21e2c-4858-4bbe-bdd8-b3c7dbe97d8b',
            '110b340b-9bbd-4836-9a7c-7b9f38ba4440',
            'ed91bd3e-dae4-4526-b0d5-5336853ed22b',
            '1cf651a4-1fc6-4b43-aefd-a213fc091bfb',
            '4376fa8c-ddcf-4931-95b9-916d090c7129',
            '90c134aa-98f0-4547-8eaf-d76d268cf662',
            'cb772ade-3a63-4940-9eac-926b7fab94b0'
        )
    ) d
    WHERE o.id = d.id;
END
$demo$;
