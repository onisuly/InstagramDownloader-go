package main

type UserMedia struct {
	Data struct {
		User struct {
			EdgeOwnerToTimelineMedia struct {
				Count    int `json:"count"`
				PageInfo struct {
					HasNextPage bool   `json:"has_next_page"`
					EndCursor   string `json:"end_cursor"`
				} `json:"page_info"`
				Edges []struct {
					Node struct {
						Typename              string `json:"__typename"`
						ID                    string `json:"id"`
						DisplayURL            string `json:"display_url"`
						IsVideo               bool   `json:"is_video"`
						VideoURL              string `json:"video_url"`
						TakenAtTimestamp      int    `json:"taken_at_timestamp"`
						EdgeSidecarToChildren struct {
							Edges []struct {
								Node struct {
									Typename   string `json:"__typename"`
									ID         string `json:"id"`
									DisplayURL string `json:"display_url"`
									IsVideo    bool   `json:"is_video"`
								} `json:"node"`
							} `json:"edges"`
						} `json:"edge_sidecar_to_children"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"edge_owner_to_timeline_media"`
		} `json:"user"`
	} `json:"data"`
	Status string `json:"status"`
}

type UserMainPage struct {
	EntryData struct {
		ProfilePage []struct {
			Graphql struct {
				User struct {
					ID string `json:"id"`
				} `json:"user"`
			} `json:"graphql"`
		} `json:"ProfilePage"`
	} `json:"entry_data"`
}
