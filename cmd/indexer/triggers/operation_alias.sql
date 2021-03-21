create or replace 
	function set_operation_alias() returns trigger as
	$$
	declare
    	alias text;
	begin
		select name into alias from tzips where address = new.source and network = new.network;
		if alias != '' then	
			new.source_alias := alias;
		end if;
		select name into alias from tzips where address = new.destination and network = new.network;
		if alias != '' then
			new.destination_alias := alias;
		end if;
		select name into alias from tzips where address = new.delegate and network = new.network;
		if alias != '' then
			new.delegate_alias := alias;
		end if;
		return new;
	end;
	$$
language 'plpgsql';

drop trigger if exists operation_alias ON operations;

create trigger operation_alias
	before insert on operations
	for each row execute 
	procedure set_operation_alias();