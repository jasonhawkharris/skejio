export interface TourDate {
	id: string;
	date: string;
	city: string;
	state: string | null;
	venue: string;
	user_id: string;
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
