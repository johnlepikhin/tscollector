#!/usr/bin/perl

my $dirname;
BEGIN {
    use File::Basename;
    $dirname = dirname(__FILE__);
}

use lib $dirname;
use TSCollector;
use TSConfig;
use warnings;
use strict;

open my $fh, "/proc/interrupts" || die "cannot open /proc/interrupts: $?";

$_ = <$fh>;
chomp;
my @cpus = grep { /^CPU/ } split /\s+/, $_;

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

my $err = TSCollector::post($TSConfig::url, $TSConfig::username, $TSConfig::password, $t);

if (defined $err) {
    print "Some error occured: $err\n"
}
