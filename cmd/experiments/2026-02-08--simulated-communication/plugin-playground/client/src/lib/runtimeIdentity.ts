import { nanoid } from "nanoid";
import type { InstanceId, PackageId } from "./quickjsContracts";

export function createInstanceId(packageId: PackageId): InstanceId {
  return `${packageId}@${nanoid(8)}`;
}
