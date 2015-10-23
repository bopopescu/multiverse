resource "aws_subnet" "monitoring-a" {
  availability_zone       = "${var.zone-a}"
  cidr_block              = "10.0.46.0/24"
  map_public_ip_on_launch = false
  vpc_id                  = "${aws_vpc.tapglue.id}"


  tags {
    Name = "monitoring-a"
  }
}

resource "aws_subnet" "monitoring-b" {
  availability_zone       = "${var.zone-b}"
  cidr_block              = "10.0.47.0/24"
  map_public_ip_on_launch = false
  vpc_id                  = "${aws_vpc.tapglue.id}"

  tags {
    Name = "monitoring-b"
  }
}

resource "aws_route_table_association" "monitoring-a" {
  subnet_id      = "${aws_subnet.monitoring-a.id}"
  route_table_id = "${aws_route_table.to-nat.id}"
}

resource "aws_route_table_association" "monitoring-b" {
  subnet_id      = "${aws_subnet.monitoring-b.id}"
  route_table_id = "${aws_route_table.to-nat.id}"
}

resource "aws_instance" "monitoring0" {
  ami           = "${var.monitoring_ami}"
  instance_type = "${var.monitoring_instance_type}"
  subnet_id     = "${aws_subnet.monitoring-a.id}"

  security_groups = [
    "${aws_security_group.platform.id}",
    "${aws_security_group.private.id}",
  ]

  tags {
    Name = "monitoring0"
  }
}

resource "aws_instance" "monitoring1" {
  ami           = "${var.monitoring_ami}"
  instance_type = "${var.monitoring_instance_type}"
  subnet_id     = "${aws_subnet.monitoring-b.id}"

  security_groups = [
    "${aws_security_group.platform.id}",
    "${aws_security_group.private.id}",
  ]

  tags {
    Name = "monitoring1"
  }
}