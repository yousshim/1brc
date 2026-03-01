SELECT station_name,
       MIN(measurement) as min_measurement,
       MAX(measurement) as max_measurement,
       CAST(AVG(measurement) AS DECIMAL(8, 1)) as mean_measurement
FROM READ_CSV('sample.txt',
              header=false,
              columns={'station_name': 'VARCHAR', 'measurement': 'DECIMAL(8, 1)'},
              delim=';', parallel=true)
GROUP BY station_name;

