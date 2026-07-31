WHENEVER SQLERROR EXIT SQL.SQLCODE ROLLBACK

SET PAGESIZE 100
SET LINESIZE 180
SET FEEDBACK ON
SET VERIFY OFF

PROMPT ============================================================
PROMPT Verificacion del esquema SuperMarket SIG
PROMPT ============================================================

SELECT COUNT(*) AS total_tablas_sig
FROM user_tables
WHERE table_name LIKE 'SIG\_%' ESCAPE '\';

SELECT table_name
FROM user_tables
WHERE table_name LIKE 'SIG\_%' ESCAPE '\'
ORDER BY table_name;

SELECT COUNT(*) AS total_restricciones
FROM user_constraints
WHERE table_name LIKE 'SIG\_%' ESCAPE '\';

SELECT COUNT(*) AS total_indices
FROM user_indexes
WHERE table_name LIKE 'SIG\_%' ESCAPE '\';

SELECT table_name,
       constraint_name,
       constraint_type,
       status
FROM user_constraints
WHERE table_name LIKE 'SIG\_%' ESCAPE '\'
ORDER BY table_name,
         constraint_type,
         constraint_name;

SELECT table_name,
       index_name,
       uniqueness,
       status
FROM user_indexes
WHERE table_name LIKE 'SIG\_%' ESCAPE '\'
ORDER BY table_name,
         index_name;

PROMPT ============================================================
PROMPT La consulta total_tablas_sig debe devolver 18
PROMPT ============================================================