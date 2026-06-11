import React from 'react';
import {
  Bold,
  Italic,
  Heading,
  Code,
  Terminal,
  Quote,
  List,
  ListOrdered,
  Link2
} from 'lucide-react';
import { Button } from '@/components/ui/button';

interface MarkdownToolbarProps {
  textareaRef: React.RefObject<HTMLTextAreaElement | null>;
  value: string;
  setValue: (val: string) => void;
}

export const MarkdownToolbar: React.FC<MarkdownToolbarProps> = ({
  textareaRef,
  value,
  setValue
}) => {
  const insertFormatting = (type: 'bold' | 'italic' | 'heading' | 'code' | 'codeblock' | 'quote' | 'ul' | 'ol' | 'link') => {
    const textarea = textareaRef.current;
    if (!textarea) return;

    const start = textarea.selectionStart;
    const end = textarea.selectionEnd;
    const selectedText = value.substring(start, end);

    let replacement = '';
    let selectionOffsetStart = 0;
    let selectionOffsetEnd = 0;

    switch (type) {
      case 'bold':
        replacement = `**${selectedText || 'bold text'}**`;
        selectionOffsetStart = 2;
        selectionOffsetEnd = selectedText ? replacement.length - 2 : replacement.length - 2;
        break;
      case 'italic':
        replacement = `*${selectedText || 'italic text'}*`;
        selectionOffsetStart = 1;
        selectionOffsetEnd = selectedText ? replacement.length - 1 : replacement.length - 1;
        break;
      case 'heading':
        replacement = `\n### ${selectedText || 'Heading'}\n`;
        selectionOffsetStart = 5;
        selectionOffsetEnd = selectedText ? replacement.length - 1 : replacement.length - 1;
        break;
      case 'code':
        replacement = `\`${selectedText || 'code'}\``;
        selectionOffsetStart = 1;
        selectionOffsetEnd = selectedText ? replacement.length - 1 : replacement.length - 1;
        break;
      case 'codeblock':
        replacement = `\n\`\`\`javascript\n${selectedText || '// code here'}\n\`\`\`\n`;
        selectionOffsetStart = 15;
        selectionOffsetEnd = selectedText ? replacement.length - 5 : replacement.length - 5;
        break;
      case 'quote':
        replacement = `\n> ${selectedText || 'Quote'}\n`;
        selectionOffsetStart = 3;
        selectionOffsetEnd = selectedText ? replacement.length - 1 : replacement.length - 1;
        break;
      case 'ul':
        replacement = `\n- ${selectedText || 'list item'}`;
        selectionOffsetStart = 3;
        selectionOffsetEnd = selectedText ? replacement.length : replacement.length;
        break;
      case 'ol':
        replacement = `\n1. ${selectedText || 'list item'}`;
        selectionOffsetStart = 4;
        selectionOffsetEnd = selectedText ? replacement.length : replacement.length;
        break;
      case 'link':
        replacement = `[${selectedText || 'link text'}](https://example.com)`;
        selectionOffsetStart = 1;
        selectionOffsetEnd = selectedText ? selectedText.length + 1 : 10; // offset inside title brackets or url
        break;
    }

    const newValue = value.substring(0, start) + replacement + value.substring(end);
    setValue(newValue);

    // Focus textarea and restore cursor selection
    setTimeout(() => {
      textarea.focus();
      textarea.setSelectionRange(
        start + selectionOffsetStart,
        start + selectionOffsetEnd
      );
    }, 0);
  };

  const items = [
    { type: 'bold' as const, label: 'Bold', icon: <Bold className="size-3.5" /> },
    { type: 'italic' as const, label: 'Italic', icon: <Italic className="size-3.5" /> },
    { type: 'heading' as const, label: 'Heading', icon: <Heading className="size-3.5" /> },
    { type: 'code' as const, label: 'Code Inline', icon: <Code className="size-3.5" /> },
    { type: 'codeblock' as const, label: 'Code Block', icon: <Terminal className="size-3.5" /> },
    { type: 'quote' as const, label: 'Quote', icon: <Quote className="size-3.5" /> },
    { type: 'ul' as const, label: 'Bullet List', icon: <List className="size-3.5" /> },
    { type: 'ol' as const, label: 'Numbered List', icon: <ListOrdered className="size-3.5" /> },
    { type: 'link' as const, label: 'Insert Link', icon: <Link2 className="size-3.5" /> },
  ];

  return (
    <div className="flex flex-wrap items-center gap-1 p-1 border-b border-border bg-muted/40 w-full rounded-t-md">
      {items.map((item) => (
        <Button
          key={item.type}
          type="button"
          variant="ghost"
          size="icon"
          onClick={() => insertFormatting(item.type)}
          className="h-7 w-7 text-muted-foreground hover:text-foreground hover:bg-muted/80 rounded transition-colors"
          title={item.label}
        >
          {item.icon}
        </Button>
      ))}
    </div>
  );
};
