export namespace cloudsync {
	
	export class SyncResult {
	    success: boolean;
	    uploaded: number;
	    downloaded: number;
	    conflicts: number;
	    error_message?: string;
	    sync_time: number;
	
	    static createFrom(source: any = {}) {
	        return new SyncResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.uploaded = source["uploaded"];
	        this.downloaded = source["downloaded"];
	        this.conflicts = source["conflicts"];
	        this.error_message = source["error_message"];
	        this.sync_time = source["sync_time"];
	    }
	}

}

export namespace storage {
	
	export class AppSettings {
	    auto_start: boolean;
	    silent_mode: boolean;
	    hotkey: string;
	
	    static createFrom(source: any = {}) {
	        return new AppSettings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.auto_start = source["auto_start"];
	        this.silent_mode = source["silent_mode"];
	        this.hotkey = source["hotkey"];
	    }
	}
	export class ClipboardItem {
	    id: string;
	    type: string;
	    content: string;
	    preview: string;
	    hash: string;
	    byte_size: number;
	    width: number;
	    height: number;
	    is_pinned: boolean;
	    device_id: string;
	    sync_id: string;
	    sync_status: number;
	    version: number;
	    created_at: number;
	    updated_at: number;
	    is_deleted: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ClipboardItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.type = source["type"];
	        this.content = source["content"];
	        this.preview = source["preview"];
	        this.hash = source["hash"];
	        this.byte_size = source["byte_size"];
	        this.width = source["width"];
	        this.height = source["height"];
	        this.is_pinned = source["is_pinned"];
	        this.device_id = source["device_id"];
	        this.sync_id = source["sync_id"];
	        this.sync_status = source["sync_status"];
	        this.version = source["version"];
	        this.created_at = source["created_at"];
	        this.updated_at = source["updated_at"];
	        this.is_deleted = source["is_deleted"];
	    }
	}
	export class CloudSyncConfig {
	    enabled: boolean;
	    server_url: string;
	    api_key: string;
	    e2e_encryption: boolean;
	    auto_sync: boolean;
	    sync_interval: number;
	    last_sync_time: number;
	
	    static createFrom(source: any = {}) {
	        return new CloudSyncConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.enabled = source["enabled"];
	        this.server_url = source["server_url"];
	        this.api_key = source["api_key"];
	        this.e2e_encryption = source["e2e_encryption"];
	        this.auto_sync = source["auto_sync"];
	        this.sync_interval = source["sync_interval"];
	        this.last_sync_time = source["last_sync_time"];
	    }
	}
	export class SecuritySettings {
	    has_master_password: boolean;
	    is_locked: boolean;
	    max_history_count: number;
	
	    static createFrom(source: any = {}) {
	        return new SecuritySettings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.has_master_password = source["has_master_password"];
	        this.is_locked = source["is_locked"];
	        this.max_history_count = source["max_history_count"];
	    }
	}

}

