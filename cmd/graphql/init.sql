do
	{dlr}{dlr}
    begin        
        if exists(select from pg_catalog.pg_roles where rolname = 'graphql') then
            reassign owned by graphql to root;
            drop owned by graphql;
            drop user graphql;
        end if;

        create role graphql LOGIN PASSWORD '$POSTGRES_GRAPHQL';

        grant connect on database indexer to graphql;

        grant select on big_map_actions to graphql;
        grant select on big_map_states to graphql;
        grant select("id", "ptr", "key", "key_hash", "value", "level", "contract", "network", "timestamp", "protocol_id", "operation_hash", "operation_counter", "operation_nonce") on big_map_diffs to graphql;
        grant select on blocks to graphql;
        grant select("id", "network", "level", "timestamp", "tags", "entrypoints", "fail_strings", "annotations", "address", "manager", "delegate", "project_id", "tx_count", "last_action", "migrations_count") on contracts to graphql;
        grant select on migrations to graphql;
        grant select("id", "content_index", "network", "protocol_id", "hash", "counter", "nonce", "internal", "status", "timestamp", "level", "kind", "initiator", "source", "fee", "gas_limit", "storage_limit", "amount", "destination", "delegate", "entrypoint", "parameters", "deffated_storage", "consumed_gas", "storage_size", "paid_storage_size_diff", "allocated_destination_contract", "errors", "burned", "allocated_destination_contract_burned") on operations to graphql;
        grant select("id", "hash", "network", "start_level", "end_level", "alias", "cost_per_byte", "hard_gas_limit_per_operation", "hard_storage_limit_per_operation", "time_between_blocks") on protocols to graphql;
        grant select on tezos_domains to graphql;
        grant select on token_balances to graphql;
        grant select on token_metadata to graphql;
        grant select("id", "network", "contract", "initiator", "operation_id", "status", "timestamp", "level", "from", "to", "token_id", "amount") on transfers to graphql;
        grant select("id", "level", "timestamp", "address", "network", "extras", "name", "description", "version", "license", "homepage", "authors", "interfaces", "views", "events") on tzips to graphql;
    end;
	{dlr}{dlr}
language 'plpgsql';