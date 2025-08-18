import DOMPurify from 'dompurify';

/**
 * Safely strips HTML tags from a string using DOMPurify
 * @param html - HTML string to strip tags from
 * @returns Plain text string with HTML tags removed
 */
export const stripHtml = (html: string): string => {
  if (!html || typeof html !== 'string') {
    return '';
  }
  
  // Use DOMPurify to sanitize and strip all HTML tags
  // ALLOWED_TAGS: [] removes all tags, leaving only text content
  return DOMPurify.sanitize(html, { ALLOWED_TAGS: [] });
};

/**
 * Gets the plain text length of HTML content (excluding tags)
 * @param html - HTML string to measure
 * @returns Length of plain text content
 */
export const getTextLength = (html: string): number => {
  return stripHtml(html).length;
};

/**
 * Checks if HTML content is empty (only contains tags/whitespace)
 * @param html - HTML string to check
 * @returns True if content is effectively empty
 */
export const isHtmlEmpty = (html: string): boolean => {
  return stripHtml(html).trim().length === 0;
};

/**
 * Truncates HTML content to a specified length (based on plain text)
 * @param html - HTML string to truncate
 * @param maxLength - Maximum length of plain text
 * @param suffix - Optional suffix to add when truncated (default: '...')
 * @returns Truncated plain text
 */
export const truncateHtml = (html: string, maxLength: number, suffix = '...'): string => {
  const plainText = stripHtml(html);
  
  if (plainText.length <= maxLength) {
    return plainText;
  }
  
  return plainText.substring(0, maxLength) + suffix;
};

/**
 * Sanitizes HTML content while preserving safe formatting tags
 * @param html - HTML string to sanitize
 * @param allowedTags - Array of allowed HTML tags (default: basic formatting tags)
 * @returns Sanitized HTML string
 */
export const sanitizeHtml = (
  html: string, 
  allowedTags: string[] = ['p', 'br', 'strong', 'b', 'em', 'i', 'u', 'strike']
): string => {
  if (!html || typeof html !== 'string') {
    return '';
  }
  
  return DOMPurify.sanitize(html, { ALLOWED_TAGS: allowedTags });
};