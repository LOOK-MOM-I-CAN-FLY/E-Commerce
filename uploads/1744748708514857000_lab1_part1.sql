--1
--CREATE VIEW ivanova_olga_dmitrievna AS
--SELECT s.id_student, s.full_name, substring(s.email from 1 for 3) || '***@' || substring(s.email from position('@' in s.email) + 1) AS email, c.course_name
--FROM students s
--JOIN courses c ON c.course_id = 2
--WHERE 2 = ANY(s.course_ids);



CREATE VIEW ivanova_olga_dmitrievna AS
SELECT s.id_student, s.full_name, substring(s.email from 1 for 3) || '***@' || substring(s.email from position('@' in s.email) + 1) AS email, c.course_name
FROM students s
JOIN teachers t ON t.full_name = 'Иванова Ольга Дмитриевна'
JOIN courses c ON c.course_id = t.course_id
WHERE t.course_id = ANY(s.course_ids);


--2
CREATE VIEW student8_courses AS
SELECT c.course_id, c.course_name, split_part(t.full_name, ' ', 1) AS t_name
FROM students s
CROSS JOIN unnest(s.course_ids) AS c_id --unnest для преобразования массива в строки и мы как бы разворачиваем массив дублируя строки для курсов
-- То есть, если у студента в массиве два элемента, то результатом будет две строки — по одной для каждого элемента массива.
JOIN courses c ON c.course_id = c_id
JOIN teachers t ON t.course_id = c_id
WHERE s.id_student = 8;


--3
CREATE MATERIALIZED VIEW courses_info AS--материализованное представление хранит результат запроса на диске, 
--что позволяет быстрее получать данные, особенно если запрос сложный или данные не меняются часто.
SELECT c.course_name, t.full_name AS teacher_full_name, COUNT(s.id_student) AS total_students
FROM courses c
JOIN teachers t ON t.course_id = c.course_id
LEFT JOIN students s ON c.course_id = ANY(s.course_ids)--означает, что даже если для курса нет ни одного студента 
--(то есть условие не выполнится ни для одной строки в students), 
--информация о курсе (и преподавателе) всё равно попадёт в результат, а данные о студенте будут NULL.
GROUP BY c.course_name, t.full_name;--Пишем это для того чтобы COUNT посчитала сгрупированных студентов а не всех подряд


--4
CREATE VIEW students_info AS
SELECT id_student, split_part(full_name, ' ', 1) || ' ' || left(split_part(full_name, ' ', 2), 1) || '. ' || left(split_part(full_name, ' ', 3), 1) || '.' AS short_name, array_length(course_ids, 1) AS course_count, 
    CASE
        WHEN array_length(course_ids, 1) = 1 THEN 'Низкий'
        WHEN array_length(course_ids, 1) IN (2, 3) THEN 'Средний'
        WHEN array_length(course_ids, 1) > 3 THEN 'Высокий'
        ELSE 'Низкий'
    END AS workload
FROM students;
