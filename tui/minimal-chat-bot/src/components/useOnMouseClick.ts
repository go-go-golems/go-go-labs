import { useCallback, useEffect, useMemo, type RefObject } from 'react';
import { type DOMElement } from 'ink';

import { useMouse } from '@zenobius/ink-mouse';
import { useElementDimensions, useElementPosition, } from '@zenobius/ink-mouse';
type MousePosition = {
  x: number;
  y: number;
};

type MouseClickAction = 'press' | 'release' | null;
type MouseScrollAction = 'scrollup' | 'scrolldown' | null;
type MouseDragAction = 'dragging' | null;

type MouseAction = MouseClickAction | MouseScrollAction;

function isIntersecting({
  mouse: { x, y },
  element,
}: {
  mouse: MousePosition;
  element: { left: number; top: number; width: number; height: number };
}) {
  /**
   * for some reason the position is off by 13 and 2
   */
  const left = element.left;
  const top = element.top;
  const width = element.width;
  const height = element.height;
  const isOutsideHorizontally = x < left || x > left + width;
  const isOutsideVertically = y < top || y > top + height;

  return !isOutsideHorizontally && !isOutsideVertically;
}


function getElementDimensions(node: DOMElement | null) {
    if (!node) {
        return;
    }

    if (!node.yogaNode) {
        return;
    }

    const elementLayout = node.yogaNode.getComputedLayout();

    return {
        width: elementLayout.width,
        height: elementLayout.height
    }
}

/**
 * Get the position of the element.
 *
 */
function getElementPosition(node: DOMElement | null) {
  if (!node) {
    return null;
  }

  if (!node.yogaNode) {
    return null;
  }
  const elementLayout = node.yogaNode.getComputedLayout();

  const parent = walkParentPosition(node);

  const position = {
    left: elementLayout.left + parent.x,
    top: elementLayout.top + parent.y,
  };

  return position;
}

/**
 * Walk the parent ancestory to get the position of the element.
 *
 * Since InkNodes are relative by default and because Ink does not
 * provide precomputed x and y values, we need to walk the parent and
 * accumulate the x and y values.
 *
 * I only discovered this by debugging the getElementPosition before
 * and after wrapping the element in a Box with padding:
 *
 *  - before padding: { left: 0, top: 0, width: 10, height: 1 }
 *  - after padding: { left: 2, top: 0, width: 10, height: 1 }
 *
 * It's still a mystery why padding on a parent results in the child
 * having a different top value. `#todo`
 */
function walkParentPosition(node: DOMElement) {
  let parent = node.parentNode;
  let x = 0;
  let y = 0;

  while (parent) {
    if (!parent.yogaNode) {
      return { x, y };
    }

    const layout = parent.yogaNode.getComputedLayout();
    x += layout.left;
    y += layout.top;

    parent = parent.parentNode;
  }
  return { x, y };
}
function useOnMouseClick(
  ref: RefObject<DOMElement>,
  onChange: (event: boolean) => void,
) {
  const mouse = useMouse();
  const elementPosition = useElementPosition(ref, [onChange, ref.current]);
  const elementDimensions = useElementDimensions(ref, [onChange, ref.current]);
  const element = useMemo(() => {
    return {
      ...elementPosition,
      ...elementDimensions,
    }
  }, [
    elementPosition,
    elementDimensions
  ])

  const handler = useCallback(
    (position: MousePosition, action: MouseClickAction) => {
      onChange(
        isIntersecting({ element, mouse: position }) && action === 'press',
      );
    },
    [ref.current, onChange],
  );

  useEffect(
    function HandleIntersection() {
      const events = mouse.events;

      events.on('click', handler);
      return () => {
        events.off('click', handler);
      };
    },
    [ref.current, handler],
  );
}


function useOnMouseHover(
  ref: RefObject<DOMElement>,
  onChange: (event: boolean) => void,
) {
  const mouse = useMouse();

  const handler = useCallback((position: MousePosition) => {
    const elementPosition = getElementPosition(ref.current);
    const elementDimensions = getElementDimensions(ref.current);
    if (!elementPosition || !elementDimensions) {
      return;
    }
    const element = {
      ...elementPosition,
      ...elementDimensions,
    };

    const intersecting = isIntersecting({
      element,
      mouse: position,
    });

    onChange(intersecting);
  }, []);

  useEffect(function HandleIntersection() {
    const events = mouse.events;

    events.on('position', handler);
    return () => {
      events.off('position', handler);
    };
  }, []);
}

export { useOnMouseHover, useOnMouseClick };
export { isIntersecting, getElementPosition, getElementDimensions };