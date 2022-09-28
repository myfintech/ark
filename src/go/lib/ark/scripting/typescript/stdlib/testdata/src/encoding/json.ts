import { json2string } from "arksdk/encoding";

const obj = {
  key3: "k3",
  key1: "k1",
  key2: "k2",
};
export const encoded = json2string(obj);
