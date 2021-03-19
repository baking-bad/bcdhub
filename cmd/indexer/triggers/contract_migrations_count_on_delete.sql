create or replace 
	function update_contract_migrations_count_on_delete () returns trigger as
	$$
	begin
		if exists (select id from contracts c2 where address = old.address and network = old.network) THEN
		update contracts set migrations_count = migrations_count - 1
			where address = old.address and network = old.network;
		end if;
		return old;
	end;
	$$
language 'plpgsql';

drop trigger if exists contract_migrations_count ON migrations;
drop trigger if exists contract_migrations_count_on_delete ON migrations;

create trigger contract_migrations_count_on_delete
	before delete on migrations
	for each row execute 
	procedure update_contract_migrations_count_on_delete();