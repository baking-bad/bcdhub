create or replace 
	function set_big_map_updates_count() returns trigger as
    $$
	begin
		if exists (select id from big_map_states where ptr = new.ptr and network = new.network and key_hash = new.key_hash) then
		update big_map_states set count = count + 1
			where ptr = new.ptr and network = new.network and key_hash = new.key_hash;
		end if;
		return new;
	end;
	$$
language 'plpgsql';

drop trigger if exists big_map_updates_count_on_new ON big_map_diffs;

create trigger big_map_updates_count_on_new
	after insert on big_map_diffs
	for each row execute 
	procedure set_big_map_updates_count();