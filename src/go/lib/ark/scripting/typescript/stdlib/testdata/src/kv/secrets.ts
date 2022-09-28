import * as kv from "arksdk/kv";

export const secret = kv.get("secret/foo");
