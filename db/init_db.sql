
    ---- Prepare stage
    CREATE OR REPLACE FUNCTION does_table_have_column(t_name TEXT, c_name TEXT)
        RETURNS BOOLEAN
        LANGUAGE plpgsql
    AS
    $$
    DECLARE
        column_count INT;
    BEGIN
        SELECT COUNT(t.column_name)
          INTO column_count
          FROM information_schema.columns AS t
         WHERE t.table_name=t_name AND t.column_name=c_name;
        RETURN column_count > 0;
    END;
    $$;

    CREATE OR REPLACE FUNCTION does_table_exist(t_name TEXT)
        RETURNS BOOLEAN
        LANGUAGE plpgsql
    AS
    $$
    DECLARE
        column_count INT;
    BEGIN
        SELECT COUNT(t.column_name)
          INTO column_count
          FROM information_schema.columns AS t
         WHERE t.table_name=t_name;
        RETURN column_count > 0;
    END;
    $$;

    DO $$
    DECLARE
        ----------
    BEGIN
        IF does_table_have_column('solana_block', 'slot') THEN
            ALTER TABLE solana_block RENAME TO oldv1_solana_blocks;
        END IF;

        IF does_table_have_column('neon_transactions', 'neon_sign') THEN
            ALTER TABLE neon_transactions RENAME TO oldv1_neon_transactions;
        END IF;

        IF does_table_have_column('neon_transaction_logs', 'blocknumber') THEN
            ALTER TABLE neon_transaction_logs RENAME TO oldv1_neon_transaction_logs;
        END IF;

        IF does_table_have_column('solana_neon_transactions', 'sol_sign') THEN
