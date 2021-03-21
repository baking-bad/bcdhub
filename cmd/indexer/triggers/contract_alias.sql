create or replace 
	function set_contract_alias() returns trigger as
	$$
	declare
    	alias text;
	begin
		select name into alias from tzips where address = new.address and network = new.network;
		if alias != '' then	
			new.alias := alias;
		end if;
		select name into alias from tzips where address = new.delegate and network = new.network;
		if alias != '' then
			new.delegate_alias := alias;
		end if;
		return new;
	end;
	$$
language 'plpgsql';

drop trigger if exists contract_alias ON contracts;

create trigger contract_alias
	before insert on contracts
	for each row execute 
	procedure set_contract_alias();