WHENEVER SQLERROR EXIT SQL.SQLCODE ROLLBACK

SET SERVEROUTPUT ON
SET PAGESIZE 100
SET LINESIZE 180
SET FEEDBACK ON
SET VERIFY OFF

PROMPT ============================================================
PROMPT Verificacion del esquema SuperMarket SIG
PROMPT ============================================================

DECLARE
    v_total_encontradas NUMBER;
BEGIN
    SELECT COUNT(*)
    INTO v_total_encontradas
    FROM user_tables
    WHERE table_name IN (
        'SIG_SINCRONIZACION',
        'SIG_CATEGORIA',
        'SIG_EMPRESA',
        'SIG_PROVEEDOR',
        'SIG_PRODUCTO',
        'SIG_CLIENTE',
        'SIG_VENTA',
        'SIG_DETALLE_VENTA',
        'SIG_PEDIDO',
        'SIG_ENTREGA',
        'SIG_LOTE',
        'SIG_MOVIMIENTO_INV',
        'SIG_DISPOSITIVO_IOT',
        'SIG_LECTURA_IOT',
        'SIG_ALERTA',
        'SIG_INCIDENTE',
        'SIG_ACCION_CORRECTIVA',
        'SIG_AUDITORIA'
    );

    IF v_total_encontradas <> 18 THEN
        RAISE_APPLICATION_ERROR(
            -20001,
            'Esquema incompleto. Tablas encontradas: '
            || v_total_encontradas
            || ' de 18.'
        );
    END IF;

    DBMS_OUTPUT.PUT_LINE(
        'Esquema SIG verificado correctamente: 18 de 18 tablas.'
    );
END;
/

PROMPT ============================================================
PROMPT Tablas definitivas del SIG
PROMPT ============================================================

SELECT table_name
FROM user_tables
WHERE table_name IN (
    'SIG_SINCRONIZACION',
    'SIG_CATEGORIA',
    'SIG_EMPRESA',
    'SIG_PROVEEDOR',
    'SIG_PRODUCTO',
    'SIG_CLIENTE',
    'SIG_VENTA',
    'SIG_DETALLE_VENTA',
    'SIG_PEDIDO',
    'SIG_ENTREGA',
    'SIG_LOTE',
    'SIG_MOVIMIENTO_INV',
    'SIG_DISPOSITIVO_IOT',
    'SIG_LECTURA_IOT',
    'SIG_ALERTA',
    'SIG_INCIDENTE',
    'SIG_ACCION_CORRECTIVA',
    'SIG_AUDITORIA'
)
ORDER BY table_name;

PROMPT ============================================================
PROMPT Conteos
PROMPT ============================================================

SELECT COUNT(*) AS total_tablas_objetivo
FROM user_tables
WHERE table_name IN (
    'SIG_SINCRONIZACION',
    'SIG_CATEGORIA',
    'SIG_EMPRESA',
    'SIG_PROVEEDOR',
    'SIG_PRODUCTO',
    'SIG_CLIENTE',
    'SIG_VENTA',
    'SIG_DETALLE_VENTA',
    'SIG_PEDIDO',
    'SIG_ENTREGA',
    'SIG_LOTE',
    'SIG_MOVIMIENTO_INV',
    'SIG_DISPOSITIVO_IOT',
    'SIG_LECTURA_IOT',
    'SIG_ALERTA',
    'SIG_INCIDENTE',
    'SIG_ACCION_CORRECTIVA',
    'SIG_AUDITORIA'
);

SELECT COUNT(*) AS total_restricciones_objetivo
FROM user_constraints
WHERE table_name IN (
    'SIG_SINCRONIZACION',
    'SIG_CATEGORIA',
    'SIG_EMPRESA',
    'SIG_PROVEEDOR',
    'SIG_PRODUCTO',
    'SIG_CLIENTE',
    'SIG_VENTA',
    'SIG_DETALLE_VENTA',
    'SIG_PEDIDO',
    'SIG_ENTREGA',
    'SIG_LOTE',
    'SIG_MOVIMIENTO_INV',
    'SIG_DISPOSITIVO_IOT',
    'SIG_LECTURA_IOT',
    'SIG_ALERTA',
    'SIG_INCIDENTE',
    'SIG_ACCION_CORRECTIVA',
    'SIG_AUDITORIA'
);

SELECT COUNT(*) AS total_indices_objetivo
FROM user_indexes
WHERE table_name IN (
    'SIG_SINCRONIZACION',
    'SIG_CATEGORIA',
    'SIG_EMPRESA',
    'SIG_PROVEEDOR',
    'SIG_PRODUCTO',
    'SIG_CLIENTE',
    'SIG_VENTA',
    'SIG_DETALLE_VENTA',
    'SIG_PEDIDO',
    'SIG_ENTREGA',
    'SIG_LOTE',
    'SIG_MOVIMIENTO_INV',
    'SIG_DISPOSITIVO_IOT',
    'SIG_LECTURA_IOT',
    'SIG_ALERTA',
    'SIG_INCIDENTE',
    'SIG_ACCION_CORRECTIVA',
    'SIG_AUDITORIA'
);

PROMPT ============================================================
PROMPT Restricciones no habilitadas
PROMPT La consulta debe devolver cero filas
PROMPT ============================================================

SELECT table_name,
       constraint_name,
       constraint_type,
       status
FROM user_constraints
WHERE table_name IN (
    'SIG_SINCRONIZACION',
    'SIG_CATEGORIA',
    'SIG_EMPRESA',
    'SIG_PROVEEDOR',
    'SIG_PRODUCTO',
    'SIG_CLIENTE',
    'SIG_VENTA',
    'SIG_DETALLE_VENTA',
    'SIG_PEDIDO',
    'SIG_ENTREGA',
    'SIG_LOTE',
    'SIG_MOVIMIENTO_INV',
    'SIG_DISPOSITIVO_IOT',
    'SIG_LECTURA_IOT',
    'SIG_ALERTA',
    'SIG_INCIDENTE',
    'SIG_ACCION_CORRECTIVA',
    'SIG_AUDITORIA'
)
AND status <> 'ENABLED'
ORDER BY table_name,
         constraint_name;

PROMPT ============================================================
PROMPT Todas las tablas SIG existentes
PROMPT SIG_HEALTHCHECK puede aparecer como tabla provisional
PROMPT ============================================================

SELECT table_name
FROM user_tables
WHERE table_name LIKE 'SIG\_%' ESCAPE '\'
ORDER BY table_name;

PROMPT ============================================================
PROMPT Verificacion completada
PROMPT ============================================================

EXIT;