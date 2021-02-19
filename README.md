# x

Like xargs, but the way I want it to work.

I use [xargs](https://man7.org/linux/man-pages/man1/xargs.1.html) all the time and love it, but find
that it never *quite* works the way I want. I started with aliasing `x` to the defaults I wanted
(like `-t -n1 -I{}`) but quickly outgrew this.

I don't expect that I will ever replicate all of xargs in this tool, but there is definitely more to
do.

## Things that differ from xargs

1. Splitting is always done on newlines, rather than all spaces
2. Rather than having a -I parameter, we detect `{}` in the given command and replace. If none
   exists, the item is added to the end. Note that this means you can't use `{}` in the command
   naturally, but I guess I will fix that when I come to it.
