// @vitest-environment happy-dom

import { describe, expect, it } from 'vitest';
import { renderMarkdown } from './markdown';

describe('renderMarkdown with resource links', () => {
  it('resolves res- UUIDs in images', () => {
    const content = '![photo](res-123-456)';
    const html = renderMarkdown(content);

    expect(html).toContain('src="/api/v1/resources/res-123-456"');
    expect(html).toContain('alt="photo"');
  });

  it('resolves res- UUIDs in links', () => {
    const content = '[document](res-abc-def)';
    const html = renderMarkdown(content);

    expect(html).toContain('href="/api/v1/resources/res-abc-def"');
    expect(html).toContain('document');
  });

  it('handles mixed content', () => {
    const content = 'Check this ![img](res-img) and [file](res-file)';
    const html = renderMarkdown(content);

    expect(html).toContain('src="/api/v1/resources/res-img"');
    expect(html).toContain('href="/api/v1/resources/res-file"');
  });

  it('renders external mp4 links as video tags', () => {
    const content = '![demo video](https://example.com/movie.mp4)';
    const html = renderMarkdown(content);

    expect(html).toContain('<video src="https://example.com/movie.mp4"');
    expect(html).toContain('controls=""');
    expect(html).toContain('title="demo video"');
    expect(html).not.toContain('<img');
  });

  it('renders external webm links as video tags', () => {
    const content = '![clip](https://cdn.test/clip.webm)';
    const html = renderMarkdown(content);

    expect(html).toContain('<video src="https://cdn.test/clip.webm"');
    expect(html).not.toContain('<img');
  });
});
