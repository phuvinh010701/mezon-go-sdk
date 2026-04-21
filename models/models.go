// Package models defines all public data types for the Mezon Go SDK.
// All structs correspond to the Mezon API request/response shapes and are
// safe for JSON marshalling/unmarshalling.
package models

// ChannelType enumerates the supported channel kinds.
type ChannelType int

const (
	ChannelTypeChannel      ChannelType = 1
	ChannelTypeGroup        ChannelType = 2
	ChannelTypeDM           ChannelType = 3
	ChannelTypeGMeetVoice   ChannelType = 4
	ChannelTypeForum        ChannelType = 5
	ChannelTypeStreaming     ChannelType = 6
	ChannelTypeThread       ChannelType = 7
	ChannelTypeApp          ChannelType = 8
	ChannelTypeAnnouncement ChannelType = 9
	ChannelTypeMezonVoice   ChannelType = 10
)

// ChannelStreamMode enumerates stream modes used in WebSocket messages.
type ChannelStreamMode int

const (
	StreamModeChannel ChannelStreamMode = 2
	StreamModeGroup   ChannelStreamMode = 3
	StreamModeDM      ChannelStreamMode = 4
	StreamModeClan    ChannelStreamMode = 5
	StreamModeThread  ChannelStreamMode = 6
)

// MessageType enumerates message type codes sent over WebSocket.
type MessageType int

const (
	MessageTypeChat               MessageType = 0
	MessageTypeChatUpdate         MessageType = 1
	MessageTypeChatRemove         MessageType = 2
	MessageTypeTyping             MessageType = 3
	MessageTypeIndicator          MessageType = 4
	MessageTypeWelcome            MessageType = 5
	MessageTypeCreateThread       MessageType = 6
	MessageTypeCreatePin          MessageType = 7
	MessageTypeMessageBuzz        MessageType = 8
	MessageTypeTopic              MessageType = 9
	MessageTypeAuditLog           MessageType = 10
	MessageTypeSendToken          MessageType = 11
	MessageTypeEphemeral          MessageType = 12
	MessageTypeUpcomingEvent      MessageType = 13
	MessageTypeUpdateEphemeral    MessageType = 14
	MessageTypeDeleteEphemeral    MessageType = 15
	MessageTypeContact            MessageType = 16
	MessageTypeLocation           MessageType = 17
	MessageTypePoll               MessageType = 18
)

// Event enumerates the real-time event names dispatched by the SDK.
type Event string

const (
	EventChannelMessage           Event = "channel_message"
	EventMessageReaction          Event = "message_reaction"
	EventUserChannelAdded         Event = "user_channel_added"
	EventUserChannelRemoved       Event = "user_channel_removed"
	EventUserClanRemoved          Event = "user_clan_removed"
	EventChannelCreated           Event = "channel_created"
	EventChannelUpdated           Event = "channel_updated"
	EventChannelDeleted           Event = "channel_deleted"
	EventClanUpdated              Event = "clan_updated"
	EventClanProfileUpdated       Event = "clan_profile_updated_event"
	EventClanEventCreated         Event = "clan_event_created"
	EventRoleEvent                Event = "role_event"
	EventRoleAssignEvent          Event = "role_assign_event"
	EventTokenSend                Event = "token_send"
	EventGiveCoffee               Event = "give_coffee"
	EventVoiceStarted             Event = "voice_started_event"
	EventVoiceEnded               Event = "voice_ended_event"
	EventVoiceJoined              Event = "voice_joined_event"
	EventVoiceLeaved              Event = "voice_leaved_event"
	EventStreamingJoined          Event = "streaming_joined_event"
	EventStreamingLeaved          Event = "streaming_leaved_event"
	EventMessageButtonClicked     Event = "message_button_clicked"
	EventDropdownBoxSelected      Event = "dropdown_box_selected"
	EventWebRTCSignalingFwd       Event = "webrtc_signaling_fwd"
	EventNotifications            Event = "notifications"
	EventQuickMenu                Event = "quick_menu"
	EventAIAgentSessionStarted    Event = "ai_agent_session_started"
	EventAIAgentSessionEnded      Event = "ai_agent_session_ended"
	EventAIAgentSessionSummaryDone Event = "ai_agent_session_summary_done"
	EventStatusPresence           Event = "status_presence_event"
	EventStreamPresence           Event = "stream_presence_event"
)

// --------------------------------- Auth / Session ---------------------------------

// Session represents an authenticated session returned by the Mezon API.
type Session struct {
	// Token is the JWT bearer token for API calls.
	Token string `json:"token"`
	// RefreshToken is used to obtain a new token after expiry.
	RefreshToken string `json:"refresh_token"`
	// UserID is the authenticated user/bot ID.
	UserID string `json:"user_id"`
	// APIURL is the base URL for REST/Protobuf calls in this session.
	APIURL string `json:"api_url,omitempty"`
	// WSURL is the WebSocket endpoint for real-time events.
	WSURL string `json:"ws_url,omitempty"`
}

// AuthenticateRequest is the body sent to the authentication endpoint.
type AuthenticateRequest struct {
	// Account holds any sub-fields required by the auth flow (currently empty for API-key auth).
	Account map[string]any `json:"account,omitempty"`
}

// --------------------------------- Clan ---------------------------------

// ClanDesc describes a clan (server / community) the bot has access to.
type ClanDesc struct {
	// ClanID is the unique identifier of the clan.
	ClanID string `json:"clan_id"`
	// ClanName is the display name of the clan.
	ClanName string `json:"clan_name"`
	// Logo is the URL of the clan logo image.
	Logo string `json:"logo,omitempty"`
	// Banner is the URL of the clan banner image.
	Banner string `json:"banner,omitempty"`
	// CreatorID is the user ID of the clan creator.
	CreatorID string `json:"creator_id,omitempty"`
	// Status is the current status code of the clan.
	Status *int `json:"status,omitempty"`
	// BadgeCount is the number of unread badges.
	BadgeCount *int `json:"badge_count,omitempty"`
}

// ClanDescList is the response type for list-clan endpoints.
type ClanDescList struct {
	// ClanDescs is the list of clan descriptions.
	ClanDescs []ClanDesc `json:"clandesc"`
	// Cursor is the pagination cursor for the next page (empty when last page).
	Cursor string `json:"cursor,omitempty"`
}

// --------------------------------- Channel ---------------------------------

// ChannelDescription describes a channel within a clan.
type ChannelDescription struct {
	// ChannelID is the unique identifier of the channel.
	ChannelID string `json:"channel_id"`
	// ChannelLabel is the human-readable channel name.
	ChannelLabel string `json:"channel_label"`
	// Type is the channel type (see ChannelType constants).
	Type ChannelType `json:"type"`
	// ClanID is the parent clan identifier.
	ClanID string `json:"clan_id,omitempty"`
	// CreatorID is the user who created the channel.
	CreatorID string `json:"creator_id,omitempty"`
	// UserIDs is the list of user IDs with access (used for DM/Group channels).
	UserIDs []string `json:"user_ids,omitempty"`
	// CategoryID groups the channel under a category.
	CategoryID string `json:"category_id,omitempty"`
}

// ChannelDescList is the response type for list-channel endpoints.
type ChannelDescList struct {
	// ChannelDescs is the list of channel descriptions.
	ChannelDescs []ChannelDescription `json:"channeldesc"`
	// Cursor is the pagination cursor for the next page.
	Cursor string `json:"cursor,omitempty"`
}

// ChannelDetail contains detailed information about a single channel.
type ChannelDetail struct {
	// ChannelID is the unique identifier.
	ChannelID string `json:"channel_id"`
	// ChannelLabel is the display name.
	ChannelLabel string `json:"channel_label"`
	// Type is the channel type.
	Type ChannelType `json:"type"`
	// ClanID is the parent clan.
	ClanID string `json:"clan_id,omitempty"`
	// Topic is the channel topic/description.
	Topic string `json:"topic,omitempty"`
	// IsPublic indicates whether the channel is visible to all clan members.
	// Uses *bool so that false is transmitted rather than omitted.
	IsPublic *bool `json:"is_public,omitempty"`
}

// CreateChannelRequest is the request body for creating a new channel or DM.
type CreateChannelRequest struct {
	// ClanID is required for clan channels; leave empty for DMs.
	ClanID string `json:"clan_id,omitempty"`
	// Type is the channel type to create.
	Type ChannelType `json:"type"`
	// ChannelLabel is the display name for the new channel.
	ChannelLabel string `json:"channel_label,omitempty"`
	// UserIDs is used when creating DM or group channels.
	UserIDs []string `json:"user_ids,omitempty"`
	// CategoryID optionally assigns the channel to a category.
	CategoryID string `json:"category_id,omitempty"`
	// IsPublic controls visibility for clan channels.
	IsPublic *bool `json:"is_public,omitempty"`
}

// ListChannelsRequest contains filter parameters for listing channels.
type ListChannelsRequest struct {
	// ClanID filters channels by clan. Required for clan channels.
	ClanID string `json:"clan_id,omitempty"`
	// ChannelType filters by channel type.
	ChannelType *ChannelType `json:"channel_type,omitempty"`
	// Limit is the maximum number of results to return.
	Limit int `json:"limit,omitempty"`
	// Cursor is the pagination cursor from a previous response.
	Cursor string `json:"cursor,omitempty"`
	// IsMobile adjusts the response for mobile clients.
	// Uses *bool so that false is transmitted rather than omitted.
	IsMobile *bool `json:"is_mobile,omitempty"`
}

// --------------------------------- Message ---------------------------------

// MessageMention represents a user mentioned in a message.
type MessageMention struct {
	// UserID is the mentioned user's identifier.
	UserID string `json:"user_id"`
	// Username is the display name of the mentioned user.
	Username string `json:"username,omitempty"`
}

// MessageAttachment represents a file or media attached to a message.
type MessageAttachment struct {
	// URL is the publicly accessible URL of the attachment.
	URL string `json:"url"`
	// Type is the MIME type or category (e.g. "image", "document").
	Type string `json:"filetype,omitempty"`
	// Size is the file size in bytes.
	Size int64 `json:"size,omitempty"`
	// Name is the original file name.
	Name string `json:"filename,omitempty"`
}

// MessageReaction represents the aggregated reactions for a single emoji.
type MessageReaction struct {
	// EmojiID is unique identifier for the emoji.
	EmojiID string `json:"emoji_id,omitempty"`
	// Emoji is the emoji character or shortcode.
	Emoji string `json:"emoji,omitempty"`
	// SenderIDs lists the user IDs that reacted with this emoji.
	SenderIDs []string `json:"sender_ids,omitempty"`
}

// MessageRef describes a referenced (replied-to) message.
type MessageRef struct {
	// MessageRefID is the ID of the original message.
	MessageRefID string `json:"message_ref_id"`
	// MessageSenderID is the author of the original message.
	MessageSenderID string `json:"message_sender_id,omitempty"`
	// MessageSenderUsername is the display name of the original author.
	MessageSenderUsername string `json:"message_sender_username,omitempty"`
	// Content is a summary of the original message content.
	Content string `json:"content,omitempty"`
}

// MessageContent is the structured content payload of a message.
// The T field holds plain text; Embed holds rich formatted content.
type MessageContent struct {
	// T is the plain text body of the message.
	T string `json:"t,omitempty"`
	// Embed holds optional rich embed content.
	Embed *MessageEmbed `json:"embed,omitempty"`
}

// MessageEmbed is the rich content block that can appear in a message.
type MessageEmbed struct {
	// Title is the embed heading.
	Title string `json:"title,omitempty"`
	// Description is the embed body text.
	Description string `json:"description,omitempty"`
	// Thumbnail is an optional thumbnail image.
	Thumbnail *EmbedThumbnail `json:"thumbnail,omitempty"`
	// Fields is a list of name/value fields displayed in a table layout.
	Fields []EmbedField `json:"fields,omitempty"`
	// Footer is an optional footer text.
	Footer *EmbedFooter `json:"footer,omitempty"`
	// Color is the accent color as an RGB integer.
	Color *int `json:"color,omitempty"`
}

// EmbedThumbnail is the thumbnail component of a MessageEmbed.
type EmbedThumbnail struct {
	// URL is the image URL.
	URL string `json:"url"`
}

// EmbedField is a single field within a MessageEmbed.
type EmbedField struct {
	// Name is the field label.
	Name string `json:"name"`
	// Value is the field content.
	Value string `json:"value"`
	// Inline controls whether the field is rendered side-by-side with others.
	// Uses *bool so that false is transmitted rather than omitted.
	Inline *bool `json:"inline,omitempty"`
}

// EmbedFooter is the footer component of a MessageEmbed.
type EmbedFooter struct {
	// Text is the footer text.
	Text string `json:"text"`
}

// ChannelMessage is the full message envelope received from real-time events
// or returned by message-list API endpoints.
type ChannelMessage struct {
	// ID is the unique message identifier (snowflake).
	ID string `json:"id"`
	// ChannelID is the channel this message was sent to.
	ChannelID string `json:"channel_id"`
	// ClanID is the clan this message belongs to (empty for DMs).
	ClanID string `json:"clan_id,omitempty"`
	// SenderID is the user ID of the message author.
	SenderID string `json:"sender_id"`
	// Username is the display name of the message author.
	Username string `json:"username,omitempty"`
	// Content is the structured message body.
	Content *MessageContent `json:"content,omitempty"`
	// Mentions lists users mentioned in this message.
	Mentions []MessageMention `json:"mentions,omitempty"`
	// Attachments lists files attached to this message.
	Attachments []MessageAttachment `json:"attachments,omitempty"`
	// Reactions lists emoji reactions on this message.
	Reactions []MessageReaction `json:"reactions,omitempty"`
	// References lists messages this message replies to.
	References []MessageRef `json:"references,omitempty"`
	// TopicID links this message to a thread topic.
	TopicID string `json:"topic_id,omitempty"`
	// Code is the MessageType code (see MessageType constants).
	Code MessageType `json:"code"`
	// CreateTime is the message creation timestamp (RFC3339 or Unix ms string).
	CreateTime string `json:"create_time,omitempty"`
	// UpdateTime is the last edit timestamp.
	UpdateTime string `json:"update_time,omitempty"`
	// MentionEveryone indicates whether this message mentions @everyone.
	// Uses *bool so that false is transmitted rather than omitted.
	MentionEveryone *bool `json:"mention_everyone,omitempty"`
	// Anonymous indicates the sender ID is hidden.
	// Uses *bool so that false is transmitted rather than omitted.
	Anonymous *bool `json:"anonymous,omitempty"`
	// Mode is the stream mode used for this message.
	Mode ChannelStreamMode `json:"mode,omitempty"`
	// IsPublic indicates whether the channel is public.
	// Uses *bool so that false is transmitted rather than omitted.
	IsPublic *bool `json:"is_public,omitempty"`
}

// ChannelMessageAck is the acknowledgement returned after sending a message.
type ChannelMessageAck struct {
	// CID is the command ID echoed from the request.
	CID string `json:"cid,omitempty"`
	// MessageID is the newly assigned message identifier.
	MessageID string `json:"message_id,omitempty"`
	// ChannelID is the channel the message was sent to.
	ChannelID string `json:"channel_id,omitempty"`
	// Error holds any server-side error message.
	Error string `json:"error,omitempty"`
}

// SendMessageRequest is the request body for sending a new message.
type SendMessageRequest struct {
	// ClanID is the clan identifier (required for non-DM messages).
	ClanID string `json:"clan_id,omitempty"`
	// ChannelID is the target channel identifier.
	ChannelID string `json:"channel_id"`
	// Mode is the stream mode for this channel.
	Mode ChannelStreamMode `json:"mode"`
	// IsPublic must match the channel's public setting.
	// Uses *bool so that false is transmitted rather than omitted.
	IsPublic *bool `json:"is_public,omitempty"`
	// Content is the structured message body.
	Content *MessageContent `json:"content"`
	// Mentions lists users to mention.
	Mentions []MessageMention `json:"mentions,omitempty"`
	// Attachments lists files to attach.
	Attachments []MessageAttachment `json:"attachments,omitempty"`
	// References lists messages this is replying to.
	References []MessageRef `json:"references,omitempty"`
	// MentionEveryone sends an @everyone notification.
	// Uses *bool so that false is transmitted rather than omitted.
	MentionEveryone *bool `json:"mention_everyone,omitempty"`
	// Anonymous hides the sender identity.
	// Uses *bool so that false is transmitted rather than omitted.
	Anonymous *bool `json:"anonymous,omitempty"`
	// TopicID links to a thread topic.
	TopicID string `json:"topic_id,omitempty"`
	// Code is the message type (defaults to MessageTypeChat).
	Code MessageType `json:"code"`
}

// EditMessageRequest is the request body for editing an existing message.
type EditMessageRequest struct {
	// MessageID is the identifier of the message to edit.
	MessageID string `json:"message_id"`
	// ChannelID is the channel the message belongs to.
	ChannelID string `json:"channel_id"`
	// ClanID is the clan the channel belongs to.
	ClanID string `json:"clan_id,omitempty"`
	// Mode is the stream mode.
	Mode ChannelStreamMode `json:"mode"`
	// IsPublic must match the channel's public setting.
	// Uses *bool so that false is transmitted rather than omitted.
	IsPublic *bool `json:"is_public,omitempty"`
	// Content is the new message body.
	Content *MessageContent `json:"content"`
	// Mentions updates the mention list.
	Mentions []MessageMention `json:"mentions,omitempty"`
	// Attachments updates the attachment list.
	Attachments []MessageAttachment `json:"attachments,omitempty"`
	// Code is the message type.
	Code MessageType `json:"code"`
}

// DeleteMessageRequest is the request body for deleting a message.
type DeleteMessageRequest struct {
	// MessageID is the identifier of the message to delete.
	MessageID string `json:"message_id"`
	// ChannelID is the channel the message belongs to.
	ChannelID string `json:"channel_id"`
	// ClanID is the clan the channel belongs to.
	ClanID string `json:"clan_id,omitempty"`
	// Mode is the stream mode.
	Mode ChannelStreamMode `json:"mode"`
	// IsPublic must match the channel's public setting.
	// Uses *bool so that false is transmitted rather than omitted.
	IsPublic *bool `json:"is_public,omitempty"`
}

// SendEphemeralRequest is the request body for sending an ephemeral message
// visible only to specific users.
type SendEphemeralRequest struct {
	// ClanID is the clan identifier.
	ClanID string `json:"clan_id,omitempty"`
	// ChannelID is the target channel identifier.
	ChannelID string `json:"channel_id"`
	// Mode is the stream mode.
	Mode ChannelStreamMode `json:"mode"`
	// IsPublic must match the channel's public setting.
	// Uses *bool so that false is transmitted rather than omitted.
	IsPublic *bool `json:"is_public,omitempty"`
	// ReceiverIDs are the user IDs that can see this message.
	ReceiverIDs []string `json:"receiver_ids"`
	// Content is the message body.
	Content *MessageContent `json:"content"`
	// References optionally link to a previous message.
	References []MessageRef `json:"references,omitempty"`
	// Code is the message type (should be MessageTypeEphemeral).
	Code MessageType `json:"code"`
}

// AddReactionRequest is the request body for adding a reaction to a message.
type AddReactionRequest struct {
	// MessageID is the target message.
	MessageID string `json:"message_id"`
	// ChannelID is the channel the message belongs to.
	ChannelID string `json:"channel_id"`
	// ClanID is the clan the channel belongs to.
	ClanID string `json:"clan_id,omitempty"`
	// Mode is the stream mode.
	Mode ChannelStreamMode `json:"mode"`
	// IsPublic must match the channel's public setting.
	// Uses *bool so that false is transmitted rather than omitted.
	IsPublic *bool `json:"is_public,omitempty"`
	// EmojiID is the unique identifier of the emoji.
	EmojiID string `json:"emoji_id"`
	// Emoji is the emoji character or shortcode.
	Emoji string `json:"emoji"`
	// Count is the number of times to add the reaction (usually 1).
	Count int `json:"count"`
	// ActionDelete when true removes the reaction instead.
	ActionDelete bool `json:"action_delete,omitempty"`
}

// --------------------------------- User ---------------------------------

// UserInitData holds the initial user information returned by the API.
type UserInitData struct {
	// ID is the unique user identifier.
	ID string `json:"id"`
	// Username is the user's account name.
	Username string `json:"username"`
	// ClanNick is the user's display name within a clan.
	ClanNick string `json:"clan_nick,omitempty"`
	// ClanAvatar is the URL of the user's clan-specific avatar.
	ClanAvatar string `json:"clan_avatar,omitempty"`
	// DisplayName is the user's global display name.
	DisplayName string `json:"display_name,omitempty"`
	// Avatar is the URL of the user's global avatar.
	Avatar string `json:"avatar,omitempty"`
	// DMChannelID is the DM channel ID for direct messaging this user.
	DMChannelID string `json:"dm_channel_id,omitempty"`
}

// --------------------------------- Quick Menu ---------------------------------

// QuickMenuAccess describes a quick-menu entry registered by a bot.
type QuickMenuAccess struct {
	// ID is the unique identifier.
	ID string `json:"id"`
	// BotID is the bot that owns this menu entry.
	BotID string `json:"bot_id,omitempty"`
	// ChannelID is the channel this menu appears in.
	ChannelID string `json:"channel_id,omitempty"`
	// ClanID is the clan this menu appears in.
	ClanID string `json:"clan_id,omitempty"`
	// MenuName is the display label of the menu item.
	MenuName string `json:"menu_name"`
	// ActionMsg is the message sent when the menu item is triggered.
	ActionMsg string `json:"action_msg"`
	// Background is the URL of the menu's background image.
	Background string `json:"background,omitempty"`
}

// AddQuickMenuRequest is the request body for creating a quick-menu entry.
type AddQuickMenuRequest struct {
	// ChannelID is the channel where the menu appears.
	ChannelID string `json:"channel_id"`
	// ClanID is the clan where the menu appears.
	ClanID string `json:"clan_id"`
	// MenuType is the type code for the menu.
	MenuType int `json:"menu_type"`
	// ActionMsg is the action message triggered on click.
	ActionMsg string `json:"action_msg"`
	// Background is the background image URL.
	Background string `json:"background,omitempty"`
	// MenuName is the display label.
	MenuName string `json:"menu_name"`
}

// DeleteQuickMenuRequest is the request body for deleting a quick-menu entry.
type DeleteQuickMenuRequest struct {
	// ID is the quick menu entry to delete.
	ID string `json:"id"`
}

// ListQuickMenuRequest is the request body for listing quick-menu entries.
type ListQuickMenuRequest struct {
	// ChannelID filters menus by channel.
	ChannelID string `json:"channel_id,omitempty"`
	// ClanID filters menus by clan.
	ClanID string `json:"clan_id,omitempty"`
}

// QuickMenuAccessList is the response for listing quick-menu entries.
type QuickMenuAccessList struct {
	// Menus is the list of quick menu entries.
	Menus []QuickMenuAccess `json:"quick_menu_access,omitempty"`
}

// --------------------------------- Roles ---------------------------------

// RoleListEventResponse is the response for listing roles.
type RoleListEventResponse struct {
	// Roles is the list of roles returned.
	Roles []Role `json:"roles,omitempty"`
}

// Role describes a permission role within a clan.
type Role struct {
	// ID is the unique role identifier.
	ID string `json:"id"`
	// Title is the display name of the role.
	Title string `json:"title"`
	// ClanID is the clan this role belongs to.
	ClanID string `json:"clan_id,omitempty"`
	// Color is the display color as an RGB integer.
	Color *int `json:"color,omitempty"`
	// Permissions lists the permission keys granted by this role.
	Permissions []string `json:"permissions,omitempty"`
}

// UpdateRoleRequest is the request body for updating a role.
type UpdateRoleRequest struct {
	// RoleID is the role to update.
	RoleID string `json:"role_id"`
	// ClanID is the clan the role belongs to.
	ClanID string `json:"clan_id"`
	// Title is the new role name.
	Title string `json:"title,omitempty"`
	// Color is the new role color.
	Color *int `json:"color,omitempty"`
	// Permissions replaces the role's permission list.
	Permissions []string `json:"permissions,omitempty"`
	// ActivePermissionIDs sets active permissions.
	ActivePermissionIDs []string `json:"active_permission_ids,omitempty"`
}

// RoleUpdateResponse is the response for updating a role.
type RoleUpdateResponse struct {
	// RoleID is the updated role identifier.
	RoleID string `json:"role_id,omitempty"`
}

// --------------------------------- Presence ---------------------------------

// Presence describes the online/offline status of a user.
type Presence struct {
	// UserID is the user identifier.
	UserID string `json:"user_id"`
	// Status is the presence status string (e.g. "online", "offline").
	Status string `json:"status,omitempty"`
	// SessionID is the unique session identifier for this presence.
	SessionID string `json:"session_id,omitempty"`
}

// StatusPresenceEvent is the real-time event payload for presence changes.
type StatusPresenceEvent struct {
	// Joins lists users that came online.
	Joins []Presence `json:"joins,omitempty"`
	// Leaves lists users that went offline.
	Leaves []Presence `json:"leaves,omitempty"`
}

// --------------------------------- Voice ---------------------------------

// VoiceChannelUser describes a user in a voice channel.
type VoiceChannelUser struct {
	// UserID is the user identifier.
	UserID string `json:"user_id"`
	// ChannelID is the voice channel.
	ChannelID string `json:"channel_id"`
	// ClanID is the parent clan.
	ClanID string `json:"clan_id,omitempty"`
}

// VoiceChannelUserList is the response for listing users in a voice channel.
type VoiceChannelUserList struct {
	// VoiceChannelUsers is the list of voice channel users.
	VoiceChannelUsers []VoiceChannelUser `json:"voice_channel_users,omitempty"`
}

// ListVoiceUsersRequest contains parameters for listing voice channel users.
type ListVoiceUsersRequest struct {
	// ClanID is the clan to query.
	ClanID string `json:"clan_id,omitempty"`
	// ChannelID filters by a specific voice channel.
	ChannelID string `json:"channel_id,omitempty"`
}
