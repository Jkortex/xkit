export const isTypingTarget = (target: EventTarget | null): boolean => {
  const node = target as HTMLElement | null;
  if (!node) return false;
  const tag = node.tagName;

  return (
    tag === 'INPUT' ||
    tag === 'TEXTAREA' ||
    tag === 'SELECT' ||
    Boolean(node.closest('[contenteditable="true"]'))
  );
};
