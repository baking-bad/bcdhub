do
$$
    begin
        if not exists (select from pg_catalog.pg_roles where  rolname = 'graphql') then
            create role graphql LOGIN PASSWORD '$POSTGRES_GRAPHQL';
        end if;

        revoke all on database indexer from graphql;
        grant connect on database indexer to graphql;

        grant select on big_map_actions to graphql;
        grant select on big_map_diffs to graphql;
        grant select on blocks to graphql;
        grant select on contracts to graphql;
        grant select on migrations to graphql;
        grant select on operations to graphql;
        grant select on protocols to graphql;
        grant select on tezos_domains to graphql;
        grant select on token_balances to graphql;
        grant select on token_metadata to graphql;
        grant select on transfers to graphql;
        grant select on tzips to graphql;

    end
$$
language 'plpgsql';
