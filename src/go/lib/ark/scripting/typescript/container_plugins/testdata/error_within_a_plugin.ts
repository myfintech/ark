/*
This file is used to test that an error coming from this file is capture
 */
import microservice from "ark/plugins/@mantl/sre/microservice"

throw new Error("something happen")