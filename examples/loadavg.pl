#!/usr/bin/perl

use TSCollector;
use warnings;
use strict;

open my $fh, "/proc/loadavg" || die "cannot open /proc/loadavg: $?";

$_ = <$fh>;
chomp;
close $fh;

my ($la1, $la5, $la15) = $_ =~ /^([0-9.]+) ([0-9.]+) ([0-9.]+)/;

print "$la1, $la5, $la15\n";

my $t = TSCollector::newTransaction();
TSCollector::addValue($t, "loadavg1", $la1, 'FloatLast');
TSCollector::addValue($t, "loadavg5", $la5, 'FloatLast');
TSCollector::addValue($t, "loadavg15", $la15, 'FloatLast');
my $err = TSCollector::post('http://localhost:8080/add', 'test', 'test', $t);

if (defined $err) {
    print "Some error occured: $err\n"
}
