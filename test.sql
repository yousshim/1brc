WITH actual AS (
    SELECT station_name,
           CAST(min_val AS DECIMAL(8, 1)) as min_measurement,
           CAST(max_val AS DECIMAL(8, 1)) as max_measurement,
           CAST(mean_val AS DECIMAL(8, 1)) as mean_measurement
    FROM READ_CSV('actual.txt',
                  header=false,
                  columns={'station_name': 'VARCHAR', 'min_val': 'DECIMAL(8, 1)', 'max_val': 'DECIMAL(8, 1)', 'mean_val': 'DECIMAL(8, 1)'},
                  delim='/', parallel=true)
),
golden AS (
    SELECT station_name,
           MIN(measurement) as min_measurement,
           MAX(measurement) as max_measurement,
           CAST(AVG(measurement) AS DECIMAL(8, 1)) as mean_measurement
    FROM READ_CSV('sample.txt',
                  header=false,
                  columns={'station_name': 'VARCHAR', 'measurement': 'DECIMAL(8, 1)'},
                  delim=';', parallel=true)
    GROUP BY station_name
),
comparison AS (
    SELECT a.station_name,
           a.min_measurement as actual_min,
           g.min_measurement as expected_min,
           ABS(a.min_measurement - g.min_measurement) as min_diff,
           a.max_measurement as actual_max,
           g.max_measurement as expected_max,
           ABS(a.max_measurement - g.max_measurement) as max_diff,
           a.mean_measurement as actual_mean,
           g.mean_measurement as expected_mean,
           ABS(a.mean_measurement - g.mean_measurement) as mean_diff
    FROM actual a
    JOIN golden g ON a.station_name = g.station_name
)
SELECT 
    CASE 
        WHEN COUNT(*) FILTER (WHERE min_diff > 0.1 OR max_diff > 0.1 OR mean_diff > 0.1) = 0 
        THEN '✓ PASS: All measurements within tolerance of 0.1'
        ELSE '✗ FAIL: ' || COUNT(*) FILTER (WHERE min_diff > 0.1 OR max_diff > 0.1 OR mean_diff > 0.1) || ' stations exceed tolerance'
    END as result,
    COUNT(*) as total_stations
FROM comparison;
