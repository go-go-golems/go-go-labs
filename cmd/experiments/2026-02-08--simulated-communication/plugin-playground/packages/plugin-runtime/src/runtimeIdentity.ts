import { nanoid } from "nanoid";
import type { InstanceId, PackageId } from "./contracts";

export function createInstanceId(packageId: PackageId): InstanceId {
  return `${packageId}@${nanoid(8)}`;
}
