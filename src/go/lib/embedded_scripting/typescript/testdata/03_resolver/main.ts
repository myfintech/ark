import lib from './lib/lib'
// @ts-ignore
import {Sum} from 'ark/external/native/lib'
// @ts-ignore
import {Multi} from 'ark/external/external/lib'
import {example as externalExample} from './lib/lib'

const sum = Sum(3, 4)
const multi = Multi(3, 4)

export default lib
export const example = externalExample
