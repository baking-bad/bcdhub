create or replace 
	function update_contract_stats_on_new_operation() returns trigger as
	$$
	begin
		if exists (select id from contracts c2 where address = new.destination and network = new.network) THEN
		update contracts set tx_count = tx_count + 1, last_action = new.timestamp
			where address = new.destination and network = new.network;
		end if;
		return new;
	end;
	$$
language 'plpgsql';

drop trigger if exists contract_stats_on_new_operation ON operations;

create trigger contract_stats_on_new_operation
	after insert on operations
	for each row execute 
	procedure update_contract_stats_on_new_operation();