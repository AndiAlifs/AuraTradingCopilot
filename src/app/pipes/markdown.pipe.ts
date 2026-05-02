import { Pipe, PipeTransform } from '@angular/core';
import { marked } from 'marked';
import DOMPurify from 'dompurify';

@Pipe({
  name: 'markdown',
  standalone: true
})
export class MarkdownPipe implements PipeTransform {
  transform(value: string | undefined): string {
    if (!value) return '';
    const parsed = marked.parse(value);
    // Convert Promise to string if marked.parse returns Promise. 
    // In marked 4+, marked.parse can be sync if async is not true
    // DOMPurify handles string
    const htmlString = typeof parsed === 'string' ? parsed : '';
    return DOMPurify.sanitize(htmlString);
  }
}
