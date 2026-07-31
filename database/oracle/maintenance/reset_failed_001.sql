WHENEVER OSERROR EXIT FAILURE ROLLBACK
WHENEVER SQLERROR EXIT SQL.SQLCODE ROLLBACK

SET SQLBLANKLINES ON
SET SERVEROUTPUT ON
SET DEFINE OFF

DECLARE
    TYPE table_names_type IS TABLE OF VARCHAR2(30);

    table_names table_names_type := table_names_type(
        'SIG_ACCION_CORRECTIVA',
        'SIG_INCIDENTE',
        'SIG_ALERTA',
        'SIG_LECTURA_IOT',
        'SIG_DISPOSITIVO_IOT',
        'SIG_AUDITORIA',
        'SIG_MOVIMIENTO_INV',
        'SIG_LOTE',
        'SIG_ENTREGA',
        'SIG_PEDIDO',
        'SIG_DETALLE_VENTA',
        'SIG_VENTA',
        'SIG_CLIENTE',
        'SIG_PRODUCTO',
        'SIG_PROVEEDOR',
        'SIG_EMPRESA',
        'SIG_CATEGORIA',
        'SIG_SINCRONIZACION'
    );
BEGIN
    FOR index_number IN 1 .. table_names.COUNT LOOP
        BEGIN
            EXECUTE IMMEDIATE
                'DROP TABLE '
                || table_names(index_number)
                || ' CASCADE CONSTRAINTS PURGE';

            DBMS_OUTPUT.PUT_LINE(
                'Eliminada: '
                || table_names(index_number)
            );
        EXCEPTION
            WHEN OTHERS THEN
                IF SQLCODE = -942 THEN
                    DBMS_OUTPUT.PUT_LINE(
                        'No existía: '
                        || table_names(index_number)
                    );
                ELSE
                    RAISE;
                END IF;
        END;
    END LOOP;
END;
/

PROMPT Limpieza de la migracion 001 completada.

EXIT;