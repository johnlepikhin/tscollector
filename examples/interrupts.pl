#!/usr/bin/perl

use TSCollector;
use warnings;
use strict;

open my $fh, "/proc/interrupts" || die "cannot open /proc/interrupts: $?";

$_ = <$fh>;
chomp;
my @cpus = grep { /^CPU/ } split /\s+/, $_;

print "== @cpus\n";

my $t = TSCollector::newTransaction();
while (<$fh>) {
    chomp;
    $_ =~ s/^\s+//;
    my @f = split /\s+/, $_;
    $f[0] =~ s/:$//;
    my $name = $f[0] . '__' . (join '_', @f[@cpus+1..$#f]);

    for my $cpu (0..$#cpus) {
        my $v = $f[$cpu+1];
        if (defined $v && $v =~ /^\d+$/) {
            TSCollector::addValue($t, "interrupts/$name/$cpus[$cpu]", $v, 'IntLast');
        }
    }
}
close $fh;

my $err = TSCollector::post('http://localhost:8080/add', 'test', 'test', $t);

if (defined $err) {
    print "Some error occured: $err\n"
}
