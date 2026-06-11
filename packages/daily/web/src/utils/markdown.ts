import { marked } from 'marked';
import DOMPurify from 'dompurify';

const SAFE_SCHEMES = new Set(['http:', 'https:', 'mailto:', 'tel:']);
const RAW_LINK_BASE_URL = import.meta.env.VITE_MARKDOWN_LINK_BASE_URL as
  | string
  | undefined;
const LINK_BASE_URL = RAW_LINK_BASE_URL?.trim() || '';

const isRelativeLink = (href: string): boolean => {
  if (!href) return false;
  if (href.startsWith('#')) return false;
  if (/^[a-zA-Z][a-zA-Z\d+\-.]*:/.test(href)) return false;
  if (href.startsWith('//')) return false;
  return true;
};

const normalizeHref = (href: string): string => {
  const trimmed = href.trim();
  if (!trimmed) return '#';
  if (trimmed.startsWith('#')) return trimmed;

  // Resolve internal resource IDs (res-UUID)
  if (trimmed.startsWith('res-')) {
    return `/api/v1/resources/${trimmed}`;
  }

  if (isRelativeLink(trimmed)) {
    if (!LINK_BASE_URL) return '#';
    try {
      const parsed = new URL(trimmed, LINK_BASE_URL);
      return parsed.toString();
    } catch {
      return '#';
    }
  }

  try {
    const parsed = new URL(trimmed, window.location.origin);
    if (!SAFE_SCHEMES.has(parsed.protocol)) return '#';
    return parsed.toString();
  } catch {
    return '#';
  }
};

const VIDEO_EXTENSIONS = ['.mp4', '.webm', '.ogg', '.mov'];

const isVideoUrl = (url: string): boolean => {
  const lowercase = url.toLowerCase();
  return VIDEO_EXTENSIONS.some((ext) => lowercase.endsWith(ext));
};

const normalizeMediaLinks = (html: string): string => {
  if (
    !html.includes('<a') &&
    !html.includes('<img') &&
    !html.includes('<video')
  )
    return html;

  const doc = new DOMParser().parseFromString(
    `<div>${html}</div>`,
    'text/html',
  );
  const root = doc.body.firstElementChild;
  if (!root) return html;

  // Normalize <a> tags
  root.querySelectorAll('a[href]').forEach((anchor) => {
    const href = anchor.getAttribute('href') ?? '';
    const normalizedHref = normalizeHref(href);
    const relativeAndMissingBase =
      isRelativeLink(href.trim()) &&
      !LINK_BASE_URL &&
      !href.trim().startsWith('res-');

    anchor.setAttribute('href', normalizedHref);
    if (relativeAndMissingBase) {
      anchor.setAttribute('aria-disabled', 'true');
      anchor.setAttribute('data-link-state', 'missing-base');
      anchor.setAttribute(
        'title',
        '未配置文档链接基地址（VITE_MARKDOWN_LINK_BASE_URL）',
      );
      anchor.removeAttribute('target');
      anchor.removeAttribute('rel');
      return;
    }

    anchor.setAttribute('target', '_blank');
    anchor.setAttribute('rel', 'noopener noreferrer');
  });

  // Normalize <img> and <video> tags
  root.querySelectorAll('img[src], video[src]').forEach((media) => {
    const src = media.getAttribute('src') ?? '';
    const normalized = normalizeHref(src);
    media.setAttribute('src', normalized);

    // Convert <img> to <video> if it points to a video file
    if (media.tagName.toLowerCase() === 'img' && isVideoUrl(normalized)) {
      const video = doc.createElement('video');
      video.setAttribute('src', normalized);
      video.setAttribute('controls', '');
      video.setAttribute('preload', 'metadata');
      video.style.maxWidth = '100%';
      const alt = media.getAttribute('alt');
      if (alt) video.setAttribute('title', alt);
      media.replaceWith(video);
    }
  });

  return root.innerHTML;
};

/**
 * 将 Markdown 转换为安全的 HTML
 */
export const renderMarkdown = (content: string): string => {
  const rawHtml = marked.parse(content) as string;
  const safeHtml = DOMPurify.sanitize(rawHtml);
  return normalizeMediaLinks(safeHtml);
};
