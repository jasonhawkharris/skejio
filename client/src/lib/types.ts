export interface TourDate {
	id: string;
	date: string;
	city: string;
	state: string | null;
	venue: string;
	poc_name: string | null;
	poc_number: string | null;
	poc_email: string | null;
	promoter_name: string | null;
	promoter_number: string | null;
	promoter_email: string | null;
	doors: string | null;
	show_start: string | null;
	show_end: string | null;
	load_in: string | null;
	sound_check: string | null;
	advance: string | null;
	user_id: string;
	tour_id: string | null;
	rider_id: string | null;
	created_at: string;
}

export interface RepresentedUser {
	relationship_id: string;
	user_id: string;
	name: string;
	email: string;
	user_type: string;
	created_at: string;
}
