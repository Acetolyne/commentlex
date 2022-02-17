# Commentlex

### Commentlex is based off of the standard go lexer however it is modified to only return comments. It is not a complete lexer but does offer the advantage of being able to return comments based on filetype and can return comments for languages that do not use the standard // or /* */ comment syntax. Additionally it can return comments from files that use a mix of comment styles for example html files that have html comments and javascript comments.

##### Options
<u>s.Match:</u> lexer option to add additional matching on comments. For single line comments this string needs to directly follow the characters that trigger the comment ignoring any whitespaces. For multiline comments this string needs to be anywhere in the comment.

{::comment} @todo add script to automatically add supported filetypes to this readme {:/comment}

