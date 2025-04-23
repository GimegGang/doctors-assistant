WITH
    names(first_name) AS (
        SELECT * FROM (VALUES
                           ('James'), ('Mary'), ('John'), ('Patricia'), ('Robert')
                      ) AS t(first_name)
    ),
    last_names(last_name) AS (
        SELECT * FROM (VALUES
                           ('Smith'), ('Johnson'), ('Williams'), ('Brown'), ('Jones')
                      ) AS t(last_name)
    ),
    departments(department) AS (
        SELECT * FROM (VALUES
                           ('backend'), ('frontend'), ('ios'), ('android')
                      ) AS t(department)
    ),
    combinations AS (
        SELECT
            n.first_name,
            ln.last_name,
            d.department
        FROM
            names n,
            last_names ln,
            departments d,
            generate_series(1, 500)
        LIMIT 10000
    )
INSERT INTO developers (name, department, geolocation, last_known_ip, is_available)
SELECT
    first_name || ' ' || last_name AS name,
    department,
    st_makepoint(
        (-180 + random() * 360),
        (-90 + random() * 180)
    ) AS geolocation,
    ('192.168.' || floor(random() * 255)::int || '.' || floor(random() * 255)::int)::inet AS last_known_ip,
    random() > 0.5 AS is_available
FROM
    combinations;