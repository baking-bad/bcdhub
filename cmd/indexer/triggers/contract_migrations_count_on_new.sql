create or replace 
	function update_contract_migrations_count_on_new () returns trigger as
	$$
	begin
		if exists (select id from contracts c2 where address = new.address and network = new.network) THEN
		update contracts set migrations_count = migrations_count + 1
			where address = new.address and network = new.network;
		end if;
		return new;
	end;
	$$
language 'plpgsql';

drop trigger if exists contract_migrations_count_on_new ON migrations;

create trigger contract_migrations_count_on_new
	after insert on migrations
	for each row execute 
	procedure update_contract_migrations_count_on_new();