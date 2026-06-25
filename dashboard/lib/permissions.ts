export function hasPermission(permissions: string[], permission: string): boolean {
  if (permissions.includes(permission)) return true;

  const [module, action] = permission.split(":");
  if (!module || !action) return false;

  return permissions.includes(`${module}:*`);
}

export function hasAnyPermission(
  permissions: string[],
  ...targets: string[]
): boolean {
  return targets.some((t) => hasPermission(permissions, t));
}

export function hasAllPermissions(
  permissions: string[],
  ...targets: string[]
): boolean {
  return targets.every((t) => hasPermission(permissions, t));
}
